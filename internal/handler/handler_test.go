package handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/mock"
)

// Mock para las utilidades
type MockUtils struct {
	mock.Mock
}

func (m *MockUtils) ExtractMessageBody(body, messageID string) (string, error) {
	args := m.Called(body, messageID)
	return args.String(0), args.Error(1)
}

func (m *MockUtils) ValidateSQSMessage(messageBody string) (*models.SQSMessage, error) {
	args := m.Called(messageBody)
	return args.Get(0).(*models.SQSMessage), args.Error(1)
}

func (m *MockUtils) DeleteMessageFromQueue(ctx context.Context, client aws.SQSAPI, queueURL string, receiptHandle *string, messageID string) error {
	args := m.Called(ctx, client, queueURL, receiptHandle, messageID)
	return args.Error(0)
}

// Mock para el service
type MockPlantillaService struct {
	mock.Mock
}

func (m *MockPlantillaService) HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error {
	args := m.Called(ctx, msg, messageID)
	return args.Error(0)
}

// Mock para el cliente SQS
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

// Implementa el método GetQueueURL
func (m *MockSQSClient) GetQueueURL() string {
	args := m.Called()
	return args.String(0)
}

func TestHandleLambdaEvent(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient) // Crear un mock para el cliente SQS
	logger := &logs.LoggerAdapter{}

	// Configura el mock para GetQueueURL
	mockSQSClient.On("GetQueueURL").Return("http://localhost:4566/000000000000/my-queue")

	// Crear el handler con los mocks
	sqsHandler := NewSQSHandler(mockPlantillaService, mockSQSClient, mockUtils, logger)

	// Configurar el evento SQS
	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:     "1",
				Body:          `{"IDPlantilla":"123","Parametro":[]}`,
				ReceiptHandle: "receipt-handle-1",
			},
		},
	}

	// Configurar las expectativas de los mocks
	mockUtils.On("ExtractMessageBody", `{"IDPlantilla":"123","Parametro":[]}`, "1").Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On("ValidateSQSMessage", `{"IDPlantilla":"123","Parametro":[]}`).Return(&models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
	}, nil)
	mockPlantillaService.On("HandlePlantilla", mock.Anything, &models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
	}, "1").Return(nil)
	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, "http://localhost:4566/000000000000/my-queue", mock.Anything, "1").Return(nil)

	// Ejecutar el método a probar
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que no haya errores
	if err != nil {
		t.Errorf("Se esperaba un error nil, pero se obtuvo %v", err)
	}

	// Verificar que se llamaron los métodos esperados
	mockUtils.AssertExpectations(t)
	mockPlantillaService.AssertExpectations(t)
	mockSQSClient.AssertExpectations(t)

}
