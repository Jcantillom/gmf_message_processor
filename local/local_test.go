package local_test

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/connection"
	"gmf_message_processor/local"
	"gorm.io/gorm"
	"testing"
)

// MockSQSHandler simula la interfaz de SQSHandlerInterface
type MockSQSHandler struct {
	mock.Mock
}

func (m *MockSQSHandler) HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	args := m.Called(ctx, sqsEvent)
	return args.Error(0)
}

// MockDBManager simula la interfaz de DBManagerInterface
type MockDBManager struct {
	mock.Mock
}

func (m *MockDBManager) InitDB(messageID string) error {
	args := m.Called(messageID)
	return args.Error(0)
}

func (m *MockDBManager) CloseDB(messageID string) {
	m.Called(messageID)
}

func (m *MockDBManager) GetDB() *gorm.DB {
	args := m.Called()
	return args.Get(0).(*gorm.DB)
}

func TestProcessLocalEventSuccess(t *testing.T) {
	// Simular el evento SQS como archivo JSON
	mockFileContent := []byte(`
	{
		"Records": [
			{
				"messageId": "12345",
				"body": "Test Message Body"
			}
		]
	}`)

	// Crear mocks
	mockSQSHandler := new(MockSQSHandler)
	mockDBManager := new(MockDBManager)

	// Configurar el comportamiento esperado de los mocks
	mockSQSHandler.On("HandleLambdaEvent", mock.Anything, mock.Anything).Return(nil)
	mockDBManager.On("CloseDB", "").Return() // Asegurarse de configurar CloseDB

	// Ejecutar la función ProcessLocalEvent, usando el mock del archivo
	local.ProcessLocalEvent(
		mockSQSHandler,
		mockDBManager,
		func(filename string) ([]byte, error) {
			return mockFileContent, nil // Devolver el contenido simulado del archivo
		})

	// Verificar que la función HandleLambdaEvent fue llamada
	mockSQSHandler.AssertCalled(t, "HandleLambdaEvent", mock.Anything, mock.Anything)
	mockDBManager.AssertCalled(t, "CloseDB", "") // Verificar que CloseDB fue llamado
}

func TestReadSQSEventFromFileSuccess(t *testing.T) {
	// Simulamos que el archivo `test_data/event.json` contiene el evento correcto
	mockFileContent := []byte(`
	{
		"Records": [
			{
				"messageId": "12345",
				"body": "Test Message Body"
			}
		]
	}`)

	// Usamos una función anónima para simular la lectura de archivos
	mockFileReader := func(filename string) ([]byte, error) {
		return mockFileContent, nil
	}

	// Ejecutar la función
	sqsEvent, err := local.ReadSQSEventFromFile(mockFileReader)

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, sqsEvent)
	assert.Equal(t, "12345", sqsEvent.Records[0].MessageId)
}

func TestReadSQSEventFromFileFileNotFound(t *testing.T) {
	// Simulamos un error al leer el archivo
	mockFileReader := func(filename string) ([]byte, error) {
		return nil, errors.New("file not found")
	}

	// Ejecutar la función
	sqsEvent, err := local.ReadSQSEventFromFile(mockFileReader)

	// Verificar que haya error
	assert.Error(t, err)
	assert.Nil(t, sqsEvent)
}

func TestReadSQSEventFromFileInvalidJSON(t *testing.T) {
	// Simulamos que el archivo contiene JSON inválido
	mockFileReader := func(filename string) ([]byte, error) {
		return []byte(`{ invalid json `), nil
	}

	// Ejecutar la función
	sqsEvent, err := local.ReadSQSEventFromFile(mockFileReader)

	// Verificar que haya error
	assert.Error(t, err)
	assert.Nil(t, sqsEvent)
}

func TestProcessLocalEventFileReadFailure(t *testing.T) {
	// Crear mocks
	mockSQSHandler := new(MockSQSHandler)
	mockDBManager := new(MockDBManager)

	// Configurar el comportamiento esperado de los mocks
	mockDBManager.On("CloseDB", "").Return()
	mockSQSHandler.On("HandleLambdaEvent", mock.Anything, mock.Anything).Return(nil)

	// Variable para hacer tracking si CleanupApplication fue llamada
	cleanupCalled := false

	// Sobrescribir la función CleanupApplicationFunc para el test
	local.CleanupApplicationFunc = func(dbManager connection.DBManagerInterface, message string) {
		cleanupCalled = true
	}

	// Ejecutar la función ProcessLocalEvent con un fallo en la lectura del archivo
	local.ProcessLocalEvent(
		mockSQSHandler,
		mockDBManager,
		func(filename string) ([]byte, error) {
			return nil, errors.New("file not found") // Simulamos un fallo en la lectura
		})

	// Verificar que CleanupApplication fue llamada
	assert.True(t, cleanupCalled, "CleanupApplication debería haber sido llamada")
}

func TestProcessLocalEventSQSEventFailure(t *testing.T) {
	// Simular el evento SQS como archivo JSON
	mockFileContent := []byte(`
    {
        "Records": [
            {
                "messageId": "12345",
                "body": "Test Message Body"
            }
        ]
    }`)

	// Crear mocks
	mockSQSHandler := new(MockSQSHandler)
	mockDBManager := new(MockDBManager)

	// Configurar el comportamiento esperado de los mocks
	mockSQSHandler.On("HandleLambdaEvent", mock.Anything, mock.Anything).Return(errors.New("event processing error"))
	mockDBManager.On("CloseDB", "").Return()

	// Ejecutar la función ProcessLocalEvent, simulando un error en el procesamiento del evento
	local.ProcessLocalEvent(
		mockSQSHandler,
		mockDBManager,
		func(filename string) ([]byte, error) {
			return mockFileContent, nil // Devolver el contenido simulado del archivo
		})

	// Verificar que HandleLambdaEvent fue llamado y que falló
	mockSQSHandler.AssertCalled(t, "HandleLambdaEvent", mock.Anything, mock.Anything)

	// Verificar que el log de error fue registrado
	// En este caso, tendrías que asegurarte de que logs.LogError fue llamado
	// Mockear el logger o verificar manualmente si la función fue llamada

	mockDBManager.AssertCalled(t, "CloseDB", "")
}
