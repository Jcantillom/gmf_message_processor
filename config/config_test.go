package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/connection"
	awsinternal "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/repository"
	"gorm.io/gorm"
	"os"
	"testing"
)

/*
======================================================================================================
=========================================== MockPlantillaService =====================================
======================================================================================================
*/
type MockPlantillaService struct {
	mock.Mock
}

func (m *MockPlantillaService) HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error {
	args := m.Called(ctx, msg, messageID)
	return args.Error(0)
}

/*
======================================================================================================
=========================================== MockLogger ===============================================
======================================================================================================
*/
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) LogError(message string, err error, messageID string) {
	m.Called(message, err, messageID)
}

func (m *MockLogger) LogInfo(message, messageID string) {
	m.Called(message, messageID)
}

func (m *MockLogger) LogWarn(message string, messageID string, extraArgs ...string) {
	m.Called(message, messageID, extraArgs)
}

func (m *MockLogger) LogDebug(message string, messageID string) {
	m.Called(message, messageID)
}

/*
======================================================================================================
=========================================== MockDBManager ============================================
======================================================================================================
*/
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
func (m *MockDBManager) Where(query interface{}, args ...interface{}) *gorm.DB {
	return &gorm.DB{}
}

func (m *MockDBManager) First(dest interface{}, conds ...interface{}) *gorm.DB {
	return &gorm.DB{}
}

/*
======================================================================================================
=========================================== MockSecretService ========================================
======================================================================================================
*/

type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) GetSecret(secretName, messageID string) (*connection.SecretData, error) {
	args := m.Called(secretName, messageID)
	return args.Get(0).(*connection.SecretData), args.Error(1)
}

/*
======================================================================================================
=========================================== MockEmailService =========================================
======================================================================================================
*/

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(remitente, destinatarios, asunto, cuerpo string, messageID string) error {
	args := m.Called(remitente, destinatarios, asunto, cuerpo, messageID)
	return args.Error(0)
}

/*
======================================================================================================
============================================ MockSQSClient ===========================================
======================================================================================================
*/

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) DeleteMessage(
	ctx context.Context,
	input *sqs.DeleteMessageInput,
	opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) GetQueueURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockSQSClient) SendMessage(
	ctx context.Context,
	input *sqs.SendMessageInput,
	opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

/*
======================================================================================================
=========================================== MockUtilsInterface =======================================
======================================================================================================
*/

type MockUtilsInterface struct {
	mock.Mock
}

func (m *MockUtilsInterface) ExtractMessageBody(sqsBody string, messageID string) (string, error) {
	args := m.Called(sqsBody, messageID)
	return args.String(0), args.Error(1)
}

func (m *MockUtilsInterface) DeleteMessageFromQueue(
	ctx context.Context, client awsinternal.SQSAPI, queueURL string, receiptHandle *string, messageID string) error {
	args := m.Called(ctx, client, queueURL, receiptHandle, messageID)
	return args.Error(0)
}

func (m *MockUtilsInterface) ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	args := m.Called(body)
	return args.Get(0).(*models.SQSMessage), args.Error(1)
}

func (m *MockUtilsInterface) SendMessageToQueue(
	ctx context.Context, client awsinternal.SQSAPI, queueURL string, messageBody string, messageID string) error {
	args := m.Called(ctx, client, queueURL, messageBody, messageID)
	return args.Error(0)
}

/*
======================================================================================================
============================================== MockSQSHandler ========================================
======================================================================================================
*/
type MockSQSHandler struct {
	mock.Mock
}

func (m *MockSQSHandler) HandleLambdaEvent(ctx context.Context, event events.SQSEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

/*
======================================================================================================
=========================================== MockPlantillaRepository ===================================
======================================================================================================
*/
type MockPlantillaRepository struct {
	mock.Mock
}

func (m *MockPlantillaRepository) CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error) {
	args := m.Called(idPlantilla)
	return args.Bool(0), args.Get(1).(*models.Plantilla), args.Error(2)
}

const (
	smtpServer      = "smtp.test.com"
	sqsQueue        = "http://sqs-url"
	sqsURL          = "http://localhost:4566/000000000000/my-queue"
	awsRegion       = "us-east-1"
	sqsEndpoint     = "http://localhost:4566"
	logClienteNil   = "El cliente SQS no debería ser nil"
	logErrorHope    = "No se esperaba un error al inicializar el cliente SQS"
	sqsQueueURLTest = "http://localhost:4566/000000000000/test-queue"
)

// setupEnvVariables configura las variables de entorno necesarias para los tests.
func setupEnvVariables() {
	viper.Reset() // Resetea Viper para evitar conflictos en cada test
	viper.Set("APP_ENV", "development")
	viper.Set("SERVICE_ENV", "local")
	viper.Set("SECRETS_DB", "some_secret_value")
	viper.Set("SECRETS_SMTP", "another")
	viper.Set("DB_HOST", "localhost")
	viper.Set("DB_PORT", "5432")
	viper.Set("DB_NAME", "testdb")
	viper.Set("DB_SCHEMA", "public")
	viper.Set("SMTP_SERVER", smtpServer)
	viper.Set("SMTP_PORT", "587")
	viper.Set("SQS_QUEUE_URL", sqsQueue)
	viper.Set("SMTP_TIMEOUT", "30")
	viper.Set("REGION_ZONE", awsRegion)
	viper.Set("SQS_MESSAGE_DELAY", "15")
	viper.Set("MAX_RETRIES", "3")
}

func TestInitConfigMissingEnvVarTriggersFatal(t *testing.T) {
	// Crear archivo .env temporal con algunas variables faltantes
	tempEnvFile := ".env"
	content := []byte(`
				APP_ENV=development
				SERVICE_ENV=local
				SECRETS_SMTP=mysecretsmtp
				DB_HOST=localhost
				DB_PORT=5432
				DB_NAME=mydb
				DB_SCHEMA=public
				SMTP_SERVER=smtp.test.com
				SMTP_PORT=587
				SQS_QUEUE_URL=http://sqs-url
				SMTP_TIMEOUT=30
                REGION_ZONE=us-east-1
`)
	err := os.WriteFile(tempEnvFile, content, 0644)
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(tempEnvFile)
		if err != nil {
			panic(err)
		}
	}()

	// Resetear Viper antes de la prueba
	viper.Reset()
	viper.SetConfigFile(tempEnvFile)

	// Crear mock del logger
	mockLogger := new(MockLogger)
	mockLogger.On(
		"LogDebug", "Leyendo variables de entorno desde el archivo .env", "").Return()
	mockLogger.On(
		"LogError",
		"La variable de entorno SECRETS_DB no está configurada",
		mock.Anything, "").Return()

	// Crear el ConfigManager con el mock del logger y simular log.Fatalf
	manager := NewConfigManager(mockLogger)
	manager.FatalfFn = func(format string, args ...interface{}) {
		// Simular un pánico en lugar de finalizar el test
		panic(fmt.Sprintf(format, args...))
	}

	// Verificar que InitConfig produce un pánico cuando falta una variable
	assert.PanicsWithValue(t, "**** Revise la configuración de la aplicación ****", func() {
		manager.InitConfig("")
	})

	// Verificar que el logger registró los mensajes en el orden correcto
	mockLogger.AssertCalled(
		t, "LogDebug", "Leyendo variables de entorno desde el archivo .env", "")
	mockLogger.AssertCalled(
		t, "LogError",
		"La variable de entorno SECRETS_DB no está configurada",
		mock.Anything,
		"",
	)
}

func TestInitConfigLoadsEnvVariables(t *testing.T) {
	// Simular la carga del archivo .env
	viper.Reset()
	viper.SetConfigFile(".env")

	// Inicializar el mock del logger
	mockLogger := new(MockLogger)
	manager := NewConfigManager(mockLogger)

	// Configurar las expectativas para LogDebug
	mockLogger.On("LogDebug", mock.AnythingOfType("string"), mock.Anything).Return()

	// Simular todas las variables de entorno requeridas
	setupEnvVariables()

	// Ejecutar InitConfig
	manager.InitConfig("")

	// Verificar que las variables clave fueron cargadas
	assert.Equal(t, "development", viper.GetString("APP_ENV"))
	assert.Equal(t, "local", viper.GetString("SERVICE_ENV"))

	// Verificar que se llamó el log de debug
	mockLogger.AssertCalled(
		t,
		"LogDebug",
		"No se pudo cargar el archivo .env, se utilizarán las variables de entorno del sistema", "",
	)
}

func TestInitConfigEnvFileLoadedSuccessfully(t *testing.T) {
	// Simular la carga del archivo .env
	viper.Reset()
	viper.SetConfigFile(".env")

	// Inicializar el mock del logger
	mockLogger := new(MockLogger)
	manager := NewConfigManager(mockLogger)

	// Configurar las expectativas para LogDebug y LogError
	mockLogger.On("LogDebug", mock.AnythingOfType("string"), mock.Anything).Return()
	mockLogger.On(
		"LogError",
		"La variable de entorno REGION_ZONE no está configurada",
		nil, "").Return()

	// Simular todas las variables de entorno requeridas
	setupEnvVariables()

	// Ejecutar InitConfig
	manager.InitConfig("")

	// Verificar que las variables clave fueron cargadas
	assert.Equal(t, "development", viper.GetString("APP_ENV"))
	assert.Equal(t, "local", viper.GetString("SERVICE_ENV"))

	// Verificar que se llamó el log de debug
	mockLogger.AssertCalled(
		t,
		"LogDebug",
		"No se pudo cargar el archivo .env, se utilizarán las variables de entorno del sistema", "",
	)

	// No esperamos que LogError sea llamado ya que REGION_ZONE está configurada
	mockLogger.AssertNotCalled(t, "LogError")
}

func TestInitApplicationDBInitError(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "mysecretdb")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", smtpServer)
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", sqsQueue)
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_MESSAGE_DELAY", "15")
	os.Setenv("MAX_RETRIES", "3")

	// Mock the dependencies
	mockDBManager := new(MockDBManager)

	// Simular que la conexión de la base de datos falla
	mockDBManager.On("InitDB", mock.Anything).Return(errors.New("db error"))

	// Mock SecretService para manejar ambas llamadas a GetSecret
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "mysecretdb", "messageID").
		Return(&connection.SecretData{Username: "user", Password: "password"}, nil)
	mockSecretService.On("GetSecret", "another_secret_value", "messageID").
		Return(&connection.SecretData{Username: "smtpuser", Password: "smtppassword"}, nil)

	// Ejecutar InitApplication y esperar un error
	appContext, err := InitApplication(
		"messageID",
		mockSecretService,
		mockDBManager,
	)

	// Verificar que se devolvió un error
	assert.Error(t, err)
	assert.Nil(t, appContext)
	assert.EqualError(t, err, "db error")
}

func TestCleanupApplicationNoDBManager(t *testing.T) {
	// Ejecutar CleanupApplication sin DBManager
	CleanupApplication(nil, "messageID")
}

func TestCleanupApplicationDBManager(t *testing.T) {
	// Mock the dependencies
	mockDBManager := new(MockDBManager)
	mockDBManager.On("CloseDB", "messageID").Return()

	// Ejecutar CleanupApplication con un DBManager
	CleanupApplication(mockDBManager, "messageID")

	// Verificar que se llamó CloseDB
	mockDBManager.AssertCalled(t, "CloseDB", "messageID")
}

func TestLogDatabaseConnectionEstablished(t *testing.T) {
	mockLogger := new(MockLogger)
	mockLogger.On("LogDebug", "Conexión a la base de datos establecida", "testMessageID").Return()

	// Inyecta el mock del logger en tu aplicación
	logs.Logger = mockLogger

	logDatabaseConnectionEstablished("testMessageID")

	mockLogger.AssertCalled(t, "LogDebug", "Conexión a la base de datos establecida", "testMessageID")
}

func TestNewPlantillaRepository(t *testing.T) {
	mockDB := new(MockDBManager)

	// Aquí especificas qué devolver cuando se llame a GetDB
	mockDB.On("GetDB").Return(&gorm.DB{})

	repo := repository.NewPlantillaRepository(mockDB.GetDB())

	assert.NotNil(t, repo)
	assert.Equal(t, mockDB.GetDB(), repo.DB)
}

func TestInitApplicationNewPlantillaRepository(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret_value")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", smtpServer)
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", sqsQueue)
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_MESSAGE_DELAY", "15")
	os.Setenv("MAX_RETRIES", "3")

	// Crear un mock de SecretService
	mockSecretService := new(MockSecretService)

	// Configurar los mocks para SECRETS_DB y SECRETS_SMTP
	mockSecretService.On(
		"GetSecret", "some_secret_value", "testMessageID").Return(&connection.SecretData{
		Username: "dbuser",
		Password: "dbpassword",
	}, nil)

	mockSecretService.On(
		"GetSecret", "another_secret_value", "testMessageID").Return(&connection.SecretData{
		Username: "smtpuser",
		Password: "smtppassword",
	}, nil)

	// Crear un mock de DBManager
	mockDBManager := new(MockDBManager)

	// Simular que InitDB sea llamado y devuelva sin error
	mockDBManager.On("InitDB", "testMessageID").Return(nil)

	// Simular que GetDB devuelva una instancia válida de *gorm.DB
	mockDB := &gorm.DB{}
	mockDBManager.On("GetDB").Return(mockDB)

	// Inyectar los mocks en InitApplication
	appContext, err := InitApplication("testMessageID", mockSecretService, mockDBManager)

	// Verificar que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, appContext)
	assert.NotNil(t, appContext.PlantillaService)

	// Verificar que los métodos del mock hayan sido llamados correctamente
	mockSecretService.AssertCalled(t, "GetSecret", "some_secret_value", "testMessageID")
	mockSecretService.AssertCalled(t, "GetSecret", "another_secret_value", "testMessageID")
	mockDBManager.AssertCalled(t, "InitDB", "testMessageID")
	mockDBManager.AssertCalled(t, "GetDB")
}

func TestInitApplicationNewSQSHandler(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret")
	os.Setenv("SECRETS_SMTP", "another_secret")
	os.Setenv("SMTP_SERVER", smtpServer)
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", sqsQueue)
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_MESSAGE_DELAY", "15")
	os.Setenv("MAX_RETRIES", "3")

	// Crear un mock de SecretService
	mockSecretService := new(MockSecretService)

	// Configurar los mocks para SECRETS_DB y SECRETS_SMTP
	mockSecretService.On(
		"GetSecret",
		"some_secret",
		"testMessageID").Return(&connection.SecretData{
		Username: "dbuser",
		Password: "dbpassword",
	}, nil)

	mockSecretService.On(
		"GetSecret",
		"another_secret",
		"testMessageID").Return(&connection.SecretData{
		Username: "smtpuser",
		Password: "smtppassword",
	}, nil)

	// Crear un mock de DBManager
	mockDBManager := new(MockDBManager)

	// Simular que InitDB sea llamado y devuelva sin error
	mockDBManager.On("InitDB", "testMessageID").Return(nil)

	// Simular que GetDB devuelva una instancia válida de *gorm.DB
	mockDB := &gorm.DB{}

	mockDBManager.On("GetDB").Return(mockDB)

	// Crear un mock de SQSClient
	mockSQSClient := new(MockSQSClient)

	// Simular que GetQueueURL devuelva una URL válida
	mockSQSClient.On("GetQueueURL").Return(sqsQueue)

	// Crear un mock de SQSHandler
	mockSQSHandler := new(MockSQSHandler)

	// Configurar el mock de SQSHandler para que devuelva sin error
	mockSQSHandler.On("HandleLambdaEvent", mock.Anything, mock.Anything).Return(nil)

	// Inyectar los mocks en InitApplication
	appContext, err := InitApplication(
		"testMessageID",
		mockSecretService,
		mockDBManager,
	)

	// Verificar que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, appContext)
	assert.NotNil(t, appContext.PlantillaService)

}

func TestInitApplicationEmailServiceInitialization(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret_value")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", smtpServer)
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", sqsQueue)
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_MESSAGE_DELAY", "15")
	os.Setenv("MAX_RETRIES", "3")

	// Crear un mock de SecretService y DBManager
	mockSecretService := new(MockSecretService)
	mockDBManager := new(MockDBManager)

	// Configurar los mocks necesarios para que no haya errores previos
	mockSecretService.On("GetSecret", "some_secret_value", "testMessageID").
		Return(&connection.SecretData{Username: "dbuser", Password: "dbpassword"}, nil)
	mockSecretService.On("GetSecret", "another_secret_value", "testMessageID").
		Return(&connection.SecretData{Username: "smtpuser", Password: "smtppassword"}, nil)
	mockDB := &gorm.DB{}
	mockDBManager.On("InitDB", "testMessageID").Return(nil)
	mockDBManager.On("GetDB").Return(mockDB)

	// Ejecutar InitApplication
	appContext, err := InitApplication("testMessageID", mockSecretService, mockDBManager)

	// Verificar que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, appContext)

	// Verificar que el servicio de correo electrónico fue inicializado correctamente
	assert.NotNil(t, appContext.PlantillaService)
}

func TestGetSecretEnvVarNotConfigured(t *testing.T) {
	mockSecretService := new(MockSecretService)

	// Asegurar que la variable de entorno no esté configurada
	os.Unsetenv("SECRETS_DB")

	// Ejecutar getSecret y verificar que devuelve el error esperado
	_, err := getSecret(mockSecretService, "SECRETS_DB", "testMessageID")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "la variable SECRETS_DB no está configurada")
}

func TestInitializeRepository(t *testing.T) {
	mockDBManager := new(MockDBManager)

	// Simular que GetDB devuelve una instancia válida de *gorm.DB
	mockDB := &gorm.DB{}
	mockDBManager.On("GetDB").Return(mockDB)

	repo := initializeRepository(mockDBManager)

	// Verificar que el repositorio se inicializa correctamente
	assert.NotNil(t, repo)
	assert.Equal(t, mockDB, repo.DB)
}

func TestGetSecretErrorFetchingSecret(t *testing.T) {
	// Configurar la variable de entorno SECRETS_DB para que no falle la validación inicial
	os.Setenv("SECRETS_DB", "my_secret")
	defer os.Unsetenv("SECRETS_DB")

	// Crear un mock del SecretService
	mockSecretService := new(MockSecretService)

	// Configurar el mock para que falle al intentar obtener el secreto, devolviendo nil para el secreto y un error
	mockSecretService.On("GetSecret", "my_secret", "testMessageID").
		Return((*connection.SecretData)(nil), errors.New("fetch secret error"))

	// Ejecutar la función getSecret con el mock configurado
	secret, err := getSecret(mockSecretService, "SECRETS_DB", "testMessageID")

	// Verificar que se devolvió un error
	assert.Nil(t, secret)
	assert.Error(t, err)

	// Verificar que el mensaje de error sea el esperado
	assert.EqualError(t, err, "fetch secret error")

	// Verificar que se haya llamado al mock correctamente
	mockSecretService.AssertCalled(t, "GetSecret", "my_secret", "testMessageID")
}

func TestInitializeSQSClientError(t *testing.T) {
	// Configurar las variables de entorno necesarias
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_QUEUE_URL", "") // Simular una URL vacía para forzar el error
	defer func() {
		os.Unsetenv("REGION_ZONE")
		os.Unsetenv("SQS_QUEUE_URL")
	}()

	// Reiniciar Viper y asegurarse de que cargue las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Ejecutar la función de inicialización del cliente SQS
	client, err := initializeSQSClient("testMessageID")

	// Verificar que se devolvió un error
	assert.Error(t, err, "Se esperaba un error debido a la URL vacía de SQS")
	assert.Contains(t,
		err.Error(),
		"invalid queue URL",
		"El mensaje de error no contiene 'invalid queue URL'")

	// Verificar que el cliente SQS sea nil
	assert.Nil(t, client, "El cliente SQS debería ser nil en caso de error")
}

func TestInitializeSQSClientLocalstackEndpoint(t *testing.T) {
	// Configura las variables de entorno necesarias
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_QUEUE_URL", sqsURL) // Simula LocalStack
	os.Setenv("SQS_ENDPOINT", sqsEndpoint)
	defer func() {
		os.Unsetenv("REGION_ZONE")
		os.Unsetenv("SQS_QUEUE_URL")
		os.Unsetenv("SQS_ENDPOINT")
	}()

	// Reinicia Viper para asegurarse de que cargue las variables
	viper.Reset()
	viper.AutomaticEnv()

	// Llama a la función que inicializa el cliente SQS
	client, err := initializeSQSClient("testMessageID")

	// Verifica que no hubo errores
	assert.NoError(t, err, "No se esperaba un error al inicializar el cliente SQS con LocalStack")

	// Verifica que el cliente no sea nil
	assert.NotNil(t, client, logClienteNil)
}

func TestRegionNotWhenUsingLocalStack(t *testing.T) {
	// Configurar las variables de entorno necesarias para simular LocalStack
	os.Setenv("SQS_ENDPOINT", sqsEndpoint)
	os.Setenv("REGION_ZONE", awsRegion)
	os.Setenv("SQS_QUEUE_URL", sqsURL) // URL válida

	// Reiniciar Viper y asegurarse de que cargue las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Ejecutar la función de inicialización del cliente SQS
	client, err := initializeSQSClient("testMessageID")

	// Verificar que no haya error
	assert.NoError(
		t,
		err,
		"No se esperaba un error al inicializar el cliente SQS con LocalStack")

	// Verificar que el cliente SQS no sea nil
	assert.NotNil(t, client, logClienteNil)
}

func TestInitializeSQSClientDefaultRegion(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("SQS_ENDPOINT", sqsEndpoint)
	os.Setenv("SQS_QUEUE_URL", sqsQueueURLTest)
	os.Unsetenv("REGION_ZONE") // Simulamos que no se establece región explícitamente

	defer func() {
		// Limpiar después del test
		os.Unsetenv("SQS_ENDPOINT")
		os.Unsetenv("SQS_QUEUE_URL")
	}()

	// Reiniciar Viper y asegurarse de que use las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Ejecutar la inicialización del cliente SQS
	client, err := initializeSQSClient("testMessageID")

	// Comprobamos la región: si no está configurada, debería usar awsRegion
	region := viper.GetString("REGION_ZONE")
	if region == "" {
		region = awsRegion
	}
	assert.Equal(t, awsRegion, region, "La región debería ser us-east-1 por defecto")

	// Verificar que no haya errores
	assert.NoError(t, err, logErrorHope)

	// Verificar que el cliente SQS no sea nil
	assert.NotNil(t, client, logClienteNil)
}

func TestInitializeSQSClientEndpointResolver(t *testing.T) {
	// Configurar las variables de entorno necesarias
	os.Setenv("SQS_ENDPOINT", sqsEndpoint)
	os.Setenv("SQS_QUEUE_URL", sqsQueueURLTest)
	os.Unsetenv("REGION_ZONE") // Para asegurarnos de que se use la región por defecto
	defer func() {
		os.Unsetenv("SQS_ENDPOINT")
		os.Unsetenv("SQS_QUEUE_URL")
	}()

	// Reiniciar Viper y asegurarse de que use las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Ejecutar la inicialización del cliente SQS
	client, err := awsinternal.NewSQSClient(
		viper.GetString("SQS_QUEUE_URL"),
		config.LoadDefaultConfig,
	)

	// Verificar que no haya error
	assert.NoError(t, err, logErrorHope)

	// Verificar que el cliente SQS no sea nil
	assert.NotNil(t, client, logClienteNil)

	// Verificar que la URL de la cola esté configurada correctamente
	assert.Contains(
		t, client.QueueURL,
		sqsEndpoint,
		"La URL de la cola debería contener la URL de LocalStack")
}

func TestInitializeSQSClientWithLocalstackEndpoint(t *testing.T) {
	// Configurar las variables de entorno necesarias
	os.Setenv("SQS_ENDPOINT", sqsEndpoint)
	os.Setenv("REGION_ZONE", awsRegion) // Región para la firma
	os.Setenv("SQS_QUEUE_URL", sqsQueueURLTest)

	defer func() {
		os.Unsetenv("SQS_ENDPOINT")
		os.Unsetenv("REGION_ZONE")
		os.Unsetenv("SQS_QUEUE_URL")
	}()

	// Reiniciar Viper para asegurar que use las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Crear el cliente SQS utilizando la configuración simulada
	client, err := awsinternal.NewSQSClient(
		viper.GetString("SQS_QUEUE_URL"),
		func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
			options := append(optFns, config.WithRegion(awsRegion)) // Región por defecto
			return config.LoadDefaultConfig(ctx, options...)
		},
	)

	// Verificar que no haya error
	assert.NoError(t, err, logErrorHope)

	// Verificar que el cliente SQS no sea nil
	assert.NotNil(t, client, logClienteNil)

	// Verificar que la URL del endpoint se resolvió correctamente
	assert.Contains(t, client.QueueURL, sqsEndpoint, "La URL del endpoint debería ser la de LocalStack")
}
