package config

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/aws"
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
	ctx context.Context, client aws.SQSAPI, queueURL string, receiptHandle *string, messageID string) error {
	args := m.Called(ctx, client, queueURL, receiptHandle, messageID)
	return args.Error(0)
}

func (m *MockUtilsInterface) ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	args := m.Called(body)
	return args.Get(0).(*models.SQSMessage), args.Error(1)
}

func (m *MockUtilsInterface) SendMessageToQueue(
	ctx context.Context, client aws.SQSAPI, queueURL string, messageBody string, messageID string) error {
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

func TestInitConfig_MissingEnvVarTriggersFatal(t *testing.T) {
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
	mockLogger.On("LogDebug", "Leyendo variables de entorno desde el archivo .env", "").Return()
	mockLogger.On("LogError", "La variable de entorno SECRETS_DB no está configurada", mock.Anything, "").Return()

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
	mockLogger.AssertCalled(t, "LogDebug", "Leyendo variables de entorno desde el archivo .env", "")
	mockLogger.AssertCalled(t, "LogError", "La variable de entorno SECRETS_DB no está configurada", mock.Anything, "")
}

func TestInitConfig_LoadsEnvVariables(t *testing.T) {
	// Simular la carga del archivo .env
	viper.Reset()
	viper.SetConfigFile(".env")

	mockLogger := new(MockLogger)
	manager := NewConfigManager(mockLogger)

	mockLogger.On("LogDebug", mock.AnythingOfType("string"), mock.Anything).Return()

	// Simular todas las variables de entorno requeridas
	viper.Set("APP_ENV", "development")
	viper.Set("SERVICE_ENV", "local")
	viper.Set("SECRETS_DB", "some_secret_value")
	viper.Set("SECRETS_SMTP", "another")
	viper.Set("DB_HOST", "localhost")
	viper.Set("DB_PORT", "5432")
	viper.Set("DB_NAME", "testdb")
	viper.Set("DB_SCHEMA", "public")
	viper.Set("SMTP_SERVER", "smtp.test.com")
	viper.Set("SMTP_PORT", "587")
	viper.Set("SQS_QUEUE_URL", "http://sqs-url")
	viper.Set("SMTP_TIMEOUT", "30")

	// Ejecutar InitConfig
	manager.InitConfig("")

	// Verificar que las variables clave fueron cargadas
	assert.Equal(t, "development", viper.GetString("APP_ENV"))
	assert.Equal(t, "local", viper.GetString("SERVICE_ENV"))

	// Verificar que se llamó el log de debug
	mockLogger.AssertCalled(
		t,
		"LogDebug",
		"No se pudo cargar el archivo .env, se utilizarán las variables de entorno del sistema", "")
}

func TestInitConfig_EnvFileLoadedSuccessfully(t *testing.T) {
	// Crear archivo .env temporal con TODAS las variables necesarias
	tempEnvFile := ".env"
	content := []byte(`
				APP_ENV=development
				SERVICE_ENV=local
				SECRETS_DB=mysecretdb
				SECRETS_SMTP=mysecretsmtp
				DB_HOST=localhost
				DB_PORT=5432
				DB_NAME=mydb
				DB_SCHEMA=public
				SMTP_SERVER=smtp.test.com
				SMTP_PORT=587
				SQS_QUEUE_URL=http://sqs-url
				SMTP_TIMEOUT=30
`)
	err := os.WriteFile(tempEnvFile, content, 0644)
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			panic(err)
		}
	}(tempEnvFile)

	// Resetear Viper
	viper.Reset()
	viper.SetConfigFile(tempEnvFile)

	// Crear mock del logger
	mockLogger := new(MockLogger)
	mockLogger.On("LogDebug", "Leyendo variables de entorno desde el archivo .env", "").Return()

	// Crear el ConfigManager con el mock del logger
	manager := NewConfigManager(mockLogger)

	// Ejecutar InitConfig
	manager.InitConfig("")

	// Verificar que el logger registró los mensajes esperados
	mockLogger.AssertCalled(t, "LogDebug", "Leyendo variables de entorno desde el archivo .env", "")

	// Verificar que las variables clave fueron cargadas
	assert.Equal(t, "development", viper.GetString("APP_ENV"))
	assert.Equal(t, "local", viper.GetString("SERVICE_ENV"))
	assert.Equal(t, "mysecretdb", viper.GetString("SECRETS_DB"))
	assert.Equal(t, "mysecretsmtp", viper.GetString("SECRETS_SMTP"))
	assert.Equal(t, "localhost", viper.GetString("DB_HOST"))
	assert.Equal(t, "5432", viper.GetString("DB_PORT"))
	assert.Equal(t, "mydb", viper.GetString("DB_NAME"))
	assert.Equal(t, "public", viper.GetString("DB_SCHEMA"))
	assert.Equal(t, "smtp.test.com", viper.GetString("SMTP_SERVER"))
	assert.Equal(t, "587", viper.GetString("SMTP_PORT"))
	assert.Equal(t, "http://sqs-url", viper.GetString("SQS_QUEUE_URL"))
	assert.Equal(t, "30", viper.GetString("SMTP_TIMEOUT"))

}

func TestInitApplication_DBInitError(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "mysecretdb")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", "http://sqs-url")

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

func TestCleanupApplication_NoDBManager(t *testing.T) {
	// Ejecutar CleanupApplication sin DBManager
	CleanupApplication(nil, "messageID")
}

func TestCleanupApplication_DBManager(t *testing.T) {
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

func TestInitApplication_NewPlantillaRepository(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret_value")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", "http://sqs-url")

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

func TestInitApplication_NewSQSHandler(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret")
	os.Setenv("SECRETS_SMTP", "another_secret")
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", "http://sqs-url")

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
	mockSQSClient.On("GetQueueURL").Return("http://sqs-url")

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

func TestInitApplication_RepositoryInitialization(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret_value")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", "http://sqs-url")
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

	// Verificar que el repositorio fue inicializado correctamente
	assert.NotNil(t, appContext.PlantillaService)
}

func TestInitApplication_EmailServiceInitialization(t *testing.T) {
	// Establecer las variables de entorno necesarias
	os.Setenv("APP_ENV", "development")
	os.Setenv("SERVICE_ENV", "local")
	os.Setenv("SECRETS_DB", "some_secret_value")
	os.Setenv("SECRETS_SMTP", "another_secret_value")
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_TIMEOUT", "30")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "public")
	os.Setenv("SQS_QUEUE_URL", "http://sqs-url")

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
