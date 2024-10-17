package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
	"strings"
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

func (m *MockUtils) DeleteMessageFromQueue(
	ctx context.Context, client aws.SQSAPI, queueURL string, receiptHandle *string, messageID string) error {
	args := m.Called(ctx, client, queueURL, receiptHandle, messageID)
	return args.Error(0)
}

func (m *MockUtils) SendMessageToQueue(
	ctx context.Context, client aws.SQSAPI, queueURL string, messageBody string, messageID string) error {
	args := m.Called(ctx, client, queueURL, messageBody, messageID)
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

func (m *MockSQSClient) DeleteMessage(
	ctx context.Context, input *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) SendMessage(
	ctx context.Context, input *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func TestHandleLambdaEvent(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	// Asigna una URL válida de SQS
	queueURL := "http://localhost:4566/000000000000/my-queue"

	// Crear el handler con los mocks y la QueueURL
	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL, // Añadir QueueURL aquí
	)

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
	mockUtils.On(
		"ExtractMessageBody",
		`{"IDPlantilla":"123","Parametro":[]}`, "1").
		Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On(
		"ValidateSQSMessage",
		`{"IDPlantilla":"123","Parametro":[]}`).
		Return(&models.SQSMessage{
			IDPlantilla: "123",
			Parametro:   []models.ParametrosSQS{},
		}, nil)
	mockPlantillaService.On("HandlePlantilla", mock.Anything, &models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
	}, "1").Return(nil)
	mockUtils.On(
		"DeleteMessageFromQueue",
		mock.Anything,
		mockSQSClient, queueURL, // Usa la QueueURL en lugar de obtenerla del cliente
		mock.Anything, "1").Return(nil)

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

func TestHandleLambdaEvent_ErrorExtractingMessageBody(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL,
	)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:     "1",
				Body:          `{"IDPlantilla":"123","Parametro":[]}`,
				ReceiptHandle: "receipt-handle-1",
			},
		},
	}

	// Configurar el mock para `ExtractMessageBody` con error
	mockUtils.On(
		"ExtractMessageBody",
		`{"IDPlantilla":"123","Parametro":[]}`, "1",
	).Return("", fmt.Errorf("error extrayendo el cuerpo del mensaje"))

	// Configurar el mock para `DeleteMessageFromQueue` para evitar el fallo
	mockUtils.On(
		"DeleteMessageFromQueue",
		mock.Anything, mockSQSClient, queueURL, mock.Anything, "1",
	).Return(nil)

	// Ejecutar la prueba
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que ocurrió el error esperado
	if err == nil {
		t.Errorf("Se esperaba un error, pero no se obtuvo ninguno")
	}

	// Verificar que los mocks fueron llamados según lo esperado
	mockUtils.AssertExpectations(t)
}

func TestHandleLambdaEvent_ErrorValidatingMessage(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL,
	)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId:     "1",
				Body:          `{"IDPlantilla":"123","Parametro":[]}`,
				ReceiptHandle: "receipt-handle-1",
			},
		},
	}

	// Configurar el mock para `ExtractMessageBody`
	mockUtils.On(
		"ExtractMessageBody", `{"IDPlantilla":"123","Parametro":[]}`, "1",
	).Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)

	// Configurar el mock para `ValidateSQSMessage` con error
	mockUtils.On(
		"ValidateSQSMessage", `{"IDPlantilla":"123","Parametro":[]}`,
	).Return((*models.SQSMessage)(nil), fmt.Errorf("error validando el mensaje"))

	// Configurar el mock para `DeleteMessageFromQueue`
	mockUtils.On(
		"DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1",
	).Return(nil)

	// Ejecutar la prueba
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que se produjo el error esperado
	if err == nil {
		t.Errorf("Se esperaba un error, pero no se obtuvo ninguno")
	}

	// Verificar que los mocks fueron llamados según lo esperado
	mockUtils.AssertExpectations(t)
}

func TestHandleLambdaEvent_ErrorReenviandoMensaje(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	// Asigna una URL válida de SQS
	queueURL := "http://localhost:4566/000000000000/my-queue"

	// Crear el handler con los mocks
	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL, // Usa queueURL directamente
	)

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

	// Simular que `HandlePlantilla` retorna un error
	mockUtils.On(
		"ExtractMessageBody",
		`{"IDPlantilla":"123","Parametro":[]}`, "1").
		Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On(
		"ValidateSQSMessage", `{"IDPlantilla":"123","Parametro":[]}`).
		Return(&models.SQSMessage{
			IDPlantilla: "123",
			Parametro:   []models.ParametrosSQS{},
			RetryCount:  1,
		}, nil)
	mockUtils.On(
		"DeleteMessageFromQueue",
		mock.Anything, mockSQSClient,
		queueURL, // Usa queueURL directamente
		mock.Anything, "1").
		Return(nil)
	mockPlantillaService.On(
		"HandlePlantilla", mock.Anything, mock.Anything, "1").
		Return(fmt.Errorf("error procesando el mensaje"))

	// Simular un error al reenviar el mensaje a SQS
	mockUtils.On(
		"SendMessageToQueue",
		mock.Anything, mockSQSClient,
		queueURL, // Usa queueURL directamente
		mock.Anything, "1").
		Return(fmt.Errorf("error al reenviar el mensaje a SQS"))

	// Ejecutar el método a probar
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que se retornó un error
	if err == nil {
		t.Errorf("Se esperaba un error, pero se obtuvo nil")
	}

	// Verificar que se llamaron los métodos esperados
	mockUtils.AssertExpectations(t)
	mockPlantillaService.AssertExpectations(t)
	mockSQSClient.AssertExpectations(t)
}

func TestHandleLambdaEvent_RetryOnPanic(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	// Asigna una URL válida de SQS
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL, // Usa queueURL directamente
	)

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

	// Simular extracción y validación de mensaje exitosas
	mockUtils.On("ExtractMessageBody", mock.Anything, "1").Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On("ValidateSQSMessage", mock.Anything).Return(&models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
		RetryCount:  0,
	}, nil)

	// Simular la eliminación del mensaje de la cola SQS
	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(nil)

	// Forzar un panic en el servicio para probar el manejo del panic
	mockPlantillaService.On("HandlePlantilla", mock.Anything, mock.Anything, "1").Run(func(args mock.Arguments) {
		panic("simulated panic")
	}).Return(nil)

	// Simular que el reenvío del mensaje falla
	mockUtils.On("SendMessageToQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(fmt.Errorf("error al reenviar el mensaje a SQS")).Once()

	// Ejecutar el método a probar
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que no hubo error
	if err != nil {
		t.Errorf("Se esperaba que no hubiera error, pero se obtuvo: %v", err)
	}

	// Verificar que se realizaron las llamadas correctas
	mockSQSClient.AssertExpectations(t)
	mockUtils.AssertExpectations(t)
	mockPlantillaService.AssertExpectations(t)

	// Verificar que el mensaje fue reenviado tras el panic
	mockUtils.AssertCalled(t, "SendMessageToQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1")

	// Verificar que el logger registre el reintento y los errores
}

func TestHandleLambdaEvent_ErrorConvertingMessageToJSON(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	// Asigna una URL válida de SQS
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL,
	)

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

	// Simular la extracción y validación de mensaje exitosas
	mockUtils.On("ExtractMessageBody", mock.Anything, "1").Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On("ValidateSQSMessage", mock.Anything).Return(&models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
		RetryCount:  1, // Primer intento
	}, nil)

	// Simular la eliminación de la cola
	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(nil)

	// Simular el procesamiento de la plantilla
	mockPlantillaService.On("HandlePlantilla", mock.Anything, mock.Anything, "1").Return(fmt.Errorf("error procesando el mensaje"))

	// Simular un error al convertir el mensaje a JSON
	mockUtils.On("SendMessageToQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(fmt.Errorf("error al convertir destinatarios a JSON"))

	// Ejecutar el método a probar
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que se retornó un error
	if err == nil {
		t.Errorf("Se esperaba un error al convertir el mensaje a JSON, pero se obtuvo nil")
	}

	// Verificar que se llamaron los métodos esperados
	mockUtils.AssertExpectations(t)
	mockPlantillaService.AssertExpectations(t)
	mockSQSClient.AssertExpectations(t)
}

func TestHandleLambdaEvent_MaxRetriesReached(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}

	// Asigna una URL válida de SQS
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(
		mockPlantillaService,
		mockSQSClient,
		mockUtils,
		logger,
		queueURL,
	)

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

	// Simular extracción y validación de mensaje exitosas
	mockUtils.On("ExtractMessageBody", mock.Anything, "1").Return(`{"IDPlantilla":"123","Parametro":[]}`, nil)
	mockUtils.On("ValidateSQSMessage", mock.Anything).Return(&models.SQSMessage{
		IDPlantilla: "123",
		Parametro:   []models.ParametrosSQS{},
		RetryCount:  utils.GetMaxRetries(), // Alcanzamos el máximo de reintentos
	}, nil)

	// Simular eliminación de la cola
	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(nil)

	// Ejecutar el método a probar
	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	// Verificar que no hubo error
	if err != nil {
		t.Errorf("Se esperaba que no hubiera error, pero se obtuvo: %v", err)
	}

	// Verificar que el logger registre el error por máximo de reintentos
	mockUtils.AssertExpectations(t)
	mockPlantillaService.AssertExpectations(t)
	mockSQSClient.AssertExpectations(t)

	// Asegurarse de que `HandlePlantilla` no fue llamado
	mockPlantillaService.AssertNotCalled(t, "HandlePlantilla", mock.Anything, mock.Anything, mock.Anything)
}

func TestHandleLambdaEvent_ErrorDeletingMessage(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(mockPlantillaService, mockSQSClient, mockUtils, logger, queueURL)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "1", Body: "{}", ReceiptHandle: "handle-1"},
		},
	}

	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").
		Return(fmt.Errorf("delete error"))

	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	if err == nil || err.Error() != "Error eliminando mensaje de SQS: delete error" {
		t.Errorf("Expected delete error, got: %v", err)
	}
}

func TestRetryMessage_ErrorMarshalling(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(mockPlantillaService, mockSQSClient, mockUtils, logger, queueURL)

	msg := &models.SQSMessage{IDPlantilla: "123", RetryCount: 1}
	messageID := "1"

	// Sobrescribimos jsonMarshal para simular el error
	originalMarshal := jsonMarshal
	defer func() { jsonMarshal = originalMarshal }()
	jsonMarshal = func(v interface{}) ([]byte, error) {
		return nil, fmt.Errorf("marshal error")
	}

	err := sqsHandler.retryMessage(context.Background(), msg, messageID, nil)

	if err == nil || err.Error() != "Error convirtiendo mensaje a JSON: marshal error" {
		t.Errorf("Expected marshal error, got: %v", err)
	}
}

func TestHandleLambdaEvent_Success(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(mockPlantillaService, mockSQSClient, mockUtils, logger, queueURL)

	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "1", Body: "{}", ReceiptHandle: "handle-1"},
		},
	}

	mockUtils.On("DeleteMessageFromQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(nil)
	mockUtils.On("ExtractMessageBody", "{}", "1").Return("{}", nil)
	mockUtils.On("ValidateSQSMessage", "{}").Return(&models.SQSMessage{}, nil)
	mockPlantillaService.On("HandlePlantilla", mock.Anything, &models.SQSMessage{}, "1").Return(nil)

	err := sqsHandler.HandleLambdaEvent(context.Background(), sqsEvent)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestRetryMessage_Success(t *testing.T) {
	mockUtils := new(MockUtils)
	mockPlantillaService := new(MockPlantillaService)
	mockSQSClient := new(MockSQSClient)
	logger := &logs.LoggerAdapter{}
	queueURL := "http://localhost:4566/000000000000/my-queue"

	sqsHandler := NewSQSHandler(mockPlantillaService, mockSQSClient, mockUtils, logger, queueURL)

	msg := &models.SQSMessage{IDPlantilla: "123", RetryCount: 1}
	messageID := "1"

	mockUtils.On("SendMessageToQueue", mock.Anything, mockSQSClient, queueURL, mock.Anything, "1").Return(nil)

	err := sqsHandler.retryMessage(context.Background(), msg, messageID, nil)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// En handler_test.go
func TestPrintSQSEvent_ErrorMarshalling(t *testing.T) {
	// Sobrescribimos jsonMarshalIndent para simular un error
	originalMarshalIndent := jsonMarshalIndent
	defer func() { jsonMarshalIndent = originalMarshalIndent }() // Restauramos el valor original después de la prueba
	jsonMarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, fmt.Errorf("marshal indent error")
	}

	// Simular un evento SQS
	sqsEvent := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "1", Body: "{}", ReceiptHandle: "handle-1"},
		},
	}

	// Sobrescribimos logDebug para capturar el mensaje logueado
	var loggedMessage string
	originalLogDebug := logDebug
	defer func() { logDebug = originalLogDebug }()
	logDebug = func(msg, _ string) {
		loggedMessage = msg
	}

	// Ejecutar la función
	printSQSEvent(sqsEvent)

	// Verificar que el mensaje de error se haya registrado
	expectedMessage := "Error al convertir el evento a JSON: marshal indent error"
	if !strings.Contains(loggedMessage, expectedMessage) {
		t.Errorf("Expected log message to contain: %q, but got: %q", expectedMessage, loggedMessage)
	}
}
