package handler

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	awsinternal "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/service"
)

// MockSQSClient is a mock for the SQS client.
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) ReceiveMessage(
	ctx context.Context, input *sqs.ReceiveMessageInput, opts ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(
	ctx context.Context, input *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

// MockPlantillaRepository is a mock for the Plantilla repository.
type MockPlantillaRepository struct {
	mock.Mock
}

func (m *MockPlantillaRepository) CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error) {
	args := m.Called(idPlantilla)
	return args.Bool(0), args.Get(1).(*models.Plantilla), args.Error(2)
}

// MockEmailService is a mock for the Email service.
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(remitente, destinatarios, asunto, cuerpo string) error {
	args := m.Called(remitente, destinatarios, asunto, cuerpo)
	return args.Error(0)
}

func TestProcessMessage_Success(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage and DeleteMessage methods
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body: aws.String(`{
						"id_plantilla": "PC001",
						"parametros": [
							{"nombre": "nombre_archivo", "valor": "archivo1.txt"}
						]
					}`),
					ReceiptHandle: aws.String("test-receipt-handle"),
				},
			},
		}, nil)

	mockSQSClient.On("DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput")).
		Return(&sqs.DeleteMessageOutput{}, nil)

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Mock CheckPlantillaExists
	mockRepo.On(
		"CheckPlantillaExists",
		"PC001").Return(true, &models.Plantilla{
		IDPlantilla:  "PC001",
		Remitente:    "remitente@example.com",
		Destinatario: "destinatario@example.com",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Hola, &nombre_archivo.",
	}, nil)

	// Parameters to be replaced
	expectedCuerpo := "Hola, archivo1.txt."

	// Mock SendEmail
	mockEmailService.On(
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		expectedCuerpo,
	).Return(nil)

	// Call the handler to process the message
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that no error occurred
	assert.NoError(t, err)

	// Verify that the mocked methods were called with the expected values
	mockRepo.AssertCalled(t, "CheckPlantillaExists", "PC001")
	mockEmailService.AssertCalled(t,
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		expectedCuerpo,
	)
	mockSQSClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput"))
}

func TestProcessMessage_NoMessages(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage method with no messages
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{},
		}, nil)

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Execute the handler with no messages
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that no error occurred
	assert.NoError(t, err)

	// Verify that the repository or email service methods were not called
	mockRepo.AssertNotCalled(t, "CheckPlantillaExists")
	mockEmailService.AssertNotCalled(t, "SendEmail")
	mockSQSClient.AssertNotCalled(t, "DeleteMessage")
}

func TestProcessMessage_InvalidMessage(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage method with an invalid message
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body:          aws.String(`{Invalid JSON}`),
					ReceiptHandle: aws.String("test-receipt-handle"),
				},
			},
		}, nil)

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Call the handler to process the message
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that an error was received
	assert.Error(t, err)

	// Verify that the repository or email service methods were not called
	mockRepo.AssertNotCalled(t, "CheckPlantillaExists")
	mockEmailService.AssertNotCalled(t, "SendEmail")
	mockSQSClient.AssertNotCalled(t, "DeleteMessage")
}

func TestProcessMessage_ReceiveMessageError(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage method to return an error
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return((*sqs.ReceiveMessageOutput)(nil), assert.AnError) // Devolver explicitamente nil y un error

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Call the handler to process the message
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that an error occurred
	assert.Error(t, err)

	// Verify that the repository or email service methods were not called
	mockRepo.AssertNotCalled(t, "CheckPlantillaExists")
	mockEmailService.AssertNotCalled(t, "SendEmail")
	mockSQSClient.AssertNotCalled(t, "DeleteMessage")
}

func TestProcessMessage_DeleteMessageError(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage method
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body: aws.String(`{
						"id_plantilla": "PC001",
						"parametros": [
							{"nombre": "nombre_archivo", "valor": "archivo1.txt"}
						]
					}`),
					ReceiptHandle: aws.String("test-receipt-handle"),
				},
			},
		}, nil)

	// Mock CheckPlantillaExists
	mockRepo.On(
		"CheckPlantillaExists",
		"PC001").Return(true, &models.Plantilla{
		IDPlantilla:  "PC001",
		Remitente:    "remitente@example.com",
		Destinatario: "destinatario@example.com",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Hola, &nombre_archivo.",
	}, nil)

	// Mock SendEmail
	mockEmailService.On(
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		"Hola, archivo1.txt.",
	).Return(nil)

	// Mock DeleteMessage to return an empty *sqs.DeleteMessageOutput and an error
	mockSQSClient.On("DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput")).
		Return(&sqs.DeleteMessageOutput{}, assert.AnError) // Aquí devolvemos un objeto vacío y un error

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Call the handler to process the message
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that an error occurred
	assert.Error(t, err)

	// Verify that the mocked methods were called with the expected values
	mockRepo.AssertCalled(t, "CheckPlantillaExists", "PC001")
	mockEmailService.AssertCalled(t,
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		"Hola, archivo1.txt.",
	)
	mockSQSClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.AnythingOfType("*sqs.DeleteMessageInput"))
}

func TestProcessMessage_HandlePlantillaError(t *testing.T) {
	mockRepo := new(MockPlantillaRepository)
	mockEmailService := new(MockEmailService)
	mockSQSClient := new(MockSQSClient)

	plantillaService := service.NewPlantillaService(mockRepo, mockEmailService)

	// Mock the SQSClient's ReceiveMessage method
	mockSQSClient.On("ReceiveMessage", mock.Anything, mock.AnythingOfType("*sqs.ReceiveMessageInput")).
		Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body: aws.String(`{
                        "id_plantilla": "PC002",
                        "parametros": [
                            {"nombre": "nombre_archivo", "valor": "TGMF-2024082801010001.txt"},
                            {"nombre": "plataforma_origen", "valor": "STRATUS"},
                            {"nombre": "fecha_recepcion", "valor": "07/10/2024"},
                            {"nombre": "hora_recepcion", "valor": "09:19 AM"},
                            {"nombre": "codigo_rechazo", "valor": "EPCM002"},
                            {"nombre": "descripcion_rechazo", "valor": "Archivo ya existe con un estado no válido para su reproceso"}
                        ]
                    }`),
					ReceiptHandle: aws.String("test-receipt-handle"),
				},
			},
		}, nil)

	// Mock CheckPlantillaExists for the "PC002"
	mockRepo.On(
		"CheckPlantillaExists",
		"PC002").Return(true, &models.Plantilla{
		IDPlantilla:  "PC002",
		Remitente:    "remitente@example.com",
		Destinatario: "destinatario@example.com",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Hola, &nombre_archivo.",
	}, nil)

	// Mock SendEmail to return an error when trying to send the email
	mockEmailService.On(
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		"Hola, TGMF-2024082801010001.txt.",
	).Return(assert.AnError)

	sqsClient := &awsinternal.SQSClient{
		Client:   mockSQSClient,
		QueueURL: "queue-url",
	}

	sqsHandler := NewSQSHandler(plantillaService, sqsClient)

	// Call the handler to process the message
	err := sqsHandler.ProcessMessage(context.Background())

	// Verify that an error occurred (because SendEmail returns an error)
	assert.Error(t, err)

	// Verify that the mocked methods were called with the expected values
	mockRepo.AssertCalled(t, "CheckPlantillaExists", "PC002")
	mockEmailService.AssertCalled(t,
		"SendEmail",
		"remitente@example.com",
		"destinatario@example.com",
		"Asunto de prueba",
		"Hola, TGMF-2024082801010001.txt.",
	)

	// Verify that DeleteMessage was NOT called, since an error occurred before deleting the message
	mockSQSClient.AssertNotCalled(t, "DeleteMessage")
}
