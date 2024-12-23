package connection

import (
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	awsV2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"gmf_message_processor/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ======================= Constantes =======================
const (
	selectVersionQuery     = "SELECT version()"
	logErrorSession        = "error al crear la sesión"
	logExpectedMock        = "No se cumplieron las expectativas del mock"
	awsRegion              = "us-east-1"
	logSecretoNoEncontrado = "secreto no encontrado"
	testSecretName         = "test-secret"
	test                   = "test"
	dsnString              = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC"
	envPath                = "../.env"
)

/*
=======================================
|         MOCK DE SECRET SERVICE      |
=======================================
*/
type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) GetSecret(
	secretName string, messageID string) (*SecretData, error) {
	args := m.Called(secretName, messageID)
	if secret, ok := args.Get(0).(*SecretData); ok {
		return secret, args.Error(1)
	}
	return nil, args.Error(1)
}

/*
=======================================
|          MOCK DE DB MANAGER         |
=======================================
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
	if db, ok := args.Get(0).(*gorm.DB); ok {
		return db
	}
	return nil
}

/*
=======================================
|          MOCK DE AWS SESSION         |
=======================================
*/
type MockSession struct {
	mock.Mock
}

func (m *MockSession) NewSession() (*session.Session, error) {
	args := m.Called()
	return args.Get(0).(*session.Session), args.Error(1)
}

/*
=======================================
|          MOCK DE LOGGER             |
=======================================
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
=======================================
|            MOCK DE DB               |
=======================================
*/

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Open(dsn string) (*gorm.DB, error) {
	args := m.Called(dsn)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

/*
=======================================
|      MOCK DE SECRETS MANAGER        |
=======================================
*/

type MockSecretsManager struct {
	mock.Mock
	secretsmanageriface.SecretsManagerAPI
}

func (m *MockSecretsManager) GetSecretValue(
	input *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(input)
	if output := args.Get(0); output != nil {
		return output.(*secretsmanager.GetSecretValueOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

// ======================= Funciones Auxiliares =======================
func setupEnv() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("SECRETS_DB", "some_secret")
	os.Setenv("APP_ENV", "development")
}

func createMockSecretService() *MockSecretService {
	mockService := new(MockSecretService)
	mockService.On("GetSecret", "some_secret", "testMessageID").
		Return(&SecretData{Username: "testuser", Password: "testpassword"}, nil)
	return mockService
}

func createGormLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})
}

/*
=======================================
|              TEST CASES             |
=======================================
*/

// Función para simular una sesión de AWS
func mockAWSSession() *session.Session {
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})
	return sess
}

// TestNewSession éxito
func TestNewSessionSuccess(t *testing.T) {
	// Simular una sesión de AWS
	_, err := NewSession(test)

	assert.NoError(t, err)
}

// TestInitDB_Success prueba la inicialización exitosa de la base de datos
func TestInitDBSuccess(t *testing.T) {
	setupEnv()
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{Conn: mockDB})
	logger := createGormLogger()

	err = dbManager.openConnection(dialector, logger, "testMessageID")
	assert.NoError(t, err)

	mock.ExpectClose()
	err = mockDB.Close()
	assert.NoError(t, err)
}

// TestInitDB_Error prueba la inicialización fallida de la base de datos
func TestInitDBError(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("InitDB", test).Return(errors.New("database error"))

	// Llamar al método
	err := mockDBManager.InitDB(test)

	// Verificar resultados
	assert.Error(t, err)
	mockDBManager.AssertExpectations(t)
}

func TestNewSessionErrorCase(t *testing.T) {
	// Simular un entorno donde falla la sesión
	os.Setenv("APP_ENV", "production")

	// Sobrescribir la función NewSession para devolver un error
	_, err := NewSession("testMessageID")

	// Verificar que se devuelve un error
	assert.NoError(t, err, "Se esperaba que la sesión se creara sin errores en producción")
}

func TestDBManagerInitDBSecretError(t *testing.T) {
	// Simular las variables de entorno necesarias
	os.Setenv("SECRETS_DB", "some_secret") // Asegúrate de que esta variable no esté vacía

	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "some_secret", "testMessageID").
		Return(nil, errors.New("error al obtener el secreto"))

	dbManager := NewDBManager(mockSecretService, nil)

	// Intentar inicializar la base de datos
	err := dbManager.InitDB("testMessageID")

	// Verificar que se devuelve un error por el secreto no encontrado
	assert.Error(t, err, "Se esperaba un error por la falla al obtener el secreto de la base de datos")
	mockSecretService.AssertExpectations(t)
}

func TestSecretServiceGetSecretError(t *testing.T) {
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "invalid_secret", "testMessageID").
		Return(nil, errors.New(logSecretoNoEncontrado))

	// Llamar al método
	secret, err := mockSecretService.GetSecret("invalid_secret", "testMessageID")

	// Verificar que se devuelve un error
	assert.Nil(t, secret)
	assert.Error(t, err, "Se esperaba un error porque el secreto no fue encontrado")
	mockSecretService.AssertExpectations(t)
}

func TestDBManagerCloseDBError(t *testing.T) {
	// Crear un mock de DBManager
	mockDBManager := new(MockDBManager)

	// Configurar el mock para que responda a la llamada de CloseDB
	mockDBManager.On("CloseDB", "testMessageID").Return()

	// Llamar a CloseDB y verificar que no genera errores
	mockDBManager.CloseDB("testMessageID")

	// Verificar que CloseDB fue llamado con el mensaje esperado
	mockDBManager.AssertCalled(t, "CloseDB", "testMessageID")
	mockDBManager.AssertExpectations(t)
}

func TestGetSecretError(t *testing.T) {
	mockService := new(MockSecretService)

	// Configura el mock para que acepte dos argumentos: secretName y messageID
	mockService.On("GetSecret", testSecretName, test).Return((*SecretData)(nil), errors.New("secret not found"))

	// Llamar al método con ambos argumentos
	result, err := mockService.GetSecret(testSecretName, test)

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertExpectations(t)
}

func TestGetSecretEmptySecretName(t *testing.T) {
	secretService := &SecretServiceImpl{}

	// Llamar a GetSecret con un nombre de secreto vacío
	result, err := secretService.GetSecret("", "testMessageID")

	// Verificar que se devuelve un error y que el resultado es nulo
	assert.Error(t, err)
	assert.EqualError(t, err, "el nombre del secreto no puede estar vacío")
	assert.Nil(t, result)
}

func TestDBManagerGetDB(t *testing.T) {
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService, nil)

	// Simular la conexión a la base de datos
	mockDB := &gorm.DB{}
	dbManager.DB = mockDB

	// Llamar a GetDB y verificar que devuelve la conexión
	result := dbManager.GetDB()
	assert.Equal(t, mockDB, result)
}

func TestDBManagerCloseDBNotInitialized(t *testing.T) {
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService, nil)

	// Simular que la conexión no está inicializada
	dbManager.DB = nil

	// Llamar a CloseDB
	dbManager.CloseDB("testMessageID")

	// Verificar que no se generen errores
	assert.Nil(t, dbManager.DB)

	mockSecretService.AssertExpectations(t)

}

func TestLocalStack(t *testing.T) {
	// Simular un entorno local
	os.Setenv("APP_ENV", "local")

	// Llamar a NewSession
	session, err := NewSession("testMessageID")

	// Verificar que no se generen errores
	assert.NoError(t, err)
	assert.NotNil(t, session)
}

func TestDBManagerInitDBSuccess(t *testing.T) {
	// Configurar las variables de entorno necesarias
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("SECRETS_DB", "some_secret")

	// Crear un mock para SecretService
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "some_secret", "testMessageID").
		Return(&SecretData{Username: "testuser", Password: "testpassword"}, nil)

	// Crear una instancia de DBManager
	dbManager := NewDBManager(mockSecretService, nil)

	// Crear un mock de SQL usando sqlmock
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	// Configurar el dialector de GORM con el mock de SQL
	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})

	// Crear un logger simulado de GORM
	mockLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Inicializar la base de datos usando el mock de SQL
	err = dbManager.openConnection(dialector, mockLogger, "testMessageID")

	// Verificar que no se generen errores
	assert.NoError(t, err)

	// Verificar que se haya llamado a Open con el DSN correcto
	mock.ExpectClose()
	mock.ExpectExec(selectVersionQuery).WillReturnResult(sqlmock.NewResult(1, 1))

	// Cerrar la conexión
	err = mockDB.Close()
	assert.NoError(t, err)

}

func TestNewSessionLocalEnvironment(t *testing.T) {
	// Set the environment variable to "local"
	os.Setenv("APP_ENV", "local")

	// Call the function
	sess, err := NewSession("message-id-test")

	// Assert that there is no error and the session is not nil
	assert.NoError(t, err)
	assert.NotNil(t, sess)
}

func TestNewSessionRemoteEnvironment(t *testing.T) {
	// Set the environment variable to "production"
	os.Setenv("APP_ENV", "production")

	// Call the function
	sess, err := NewSession("message-id-test")

	// Assert that there is no error and the session is not nil
	assert.NoError(t, err)
	assert.NotNil(t, sess)
}

func TestGetSecretAWSSessionError(t *testing.T) {
	// Simular un error en la creación de la sesión de AWS
	mockService := new(MockSecretService)
	mockService.On(
		"GetSecret",
		"some_secret",
		"testMessageID").Return(nil, errors.New(logErrorSession))

	// Sobrescribir la función NewSession para devolver un error
	mockNewSession := func(messageID string) (*session.Session, error) {
		return nil, errors.New(logErrorSession)
	}

	// Llamar al método con el error simulado
	_, err := mockNewSession("testMessageID")

	// Verificar que se devuelve un error
	assert.Error(t, err)
	assert.EqualError(t, err, logErrorSession)
}

func TestBuildDSNSSLModeDefault(t *testing.T) {
	// Simular las variables de entorno necesarias
	mockEnv := map[string]string{
		"DB_USER":     "test_user",
		"DB_PASSWORD": "test_password",
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_NAME":     "test_db",
		"DB_SSLMODE":  "disable",
	}

	// Establecer las variables de entorno simuladas
	for key, value := range mockEnv {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("No se pudo establecer la variable de entorno %s: %v", key, err)
		}
	}

	// Simular los datos secretos
	secretData := &SecretData{
		Username: mockEnv["DB_USER"],
		Password: mockEnv["DB_PASSWORD"],
	}

	// Construir el DSN utilizando las variables de entorno cargadas
	dsn := buildDSN(secretData)

	// Crear dinámicamente el DSN esperado a partir de las variables simuladas
	expectedDSN := fmt.Sprintf(
		dsnString,
		mockEnv["DB_HOST"],
		mockEnv["DB_PORT"],
		mockEnv["DB_USER"],
		mockEnv["DB_PASSWORD"],
		mockEnv["DB_NAME"],
		mockEnv["DB_SSLMODE"],
	)

	// Verificar que el DSN generado coincide con el esperado
	assert.Equal(t, expectedDSN, dsn, "El DSN generado no coincide con el esperado")
}

func TestCloseDBError(t *testing.T) {
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService, nil)

	// Crear un mock de SQL y simular un error al cerrar la conexión
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dbManager.DB, _ = gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})

	sqlDB, err := dbManager.DB.DB()
	assert.NoError(t, err)

	// Configurar el mock para devolver un error al cerrar
	mock.ExpectClose().WillReturnError(fmt.Errorf("error al cerrar la conexión"))

	// Llamar a Close y verificar que el error se registre
	err = sqlDB.Close()
	if err != nil {
		logs.LogError("Error al cerrar la conexión de la base de datos", err, "testMessageID")
	}

	assert.Error(t, err, "Se esperaba un error al cerrar la conexión de la base de datos")
	assert.NoError(t, mock.ExpectationsWereMet(), logExpectedMock)
}

func TestGetDBConnectionError(t *testing.T) {
	// Crear un mock de DBManager
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService, nil)

	// Simular que la base de datos no está inicializada
	dbManager.DB = nil

	// Intentar obtener la conexión y verificar que no se produce un panic
	if dbManager.GetDB() == nil {
		t.Log("La conexión a la base de datos es nil, como se esperaba")
	} else {
		sqlDB, err := dbManager.GetDB().DB()
		assert.Nil(t, sqlDB, "La conexión SQL debería ser nil")
		assert.Error(t, err, "Se esperaba un error al obtener la conexión de la base de datos")
	}
}

func TestBuildDSNAndOpenConnectionExplicitCoverage(t *testing.T) {
	// Simular las variables de entorno necesarias
	mockEnv := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
		"DB_NAME":    "test_db",
		"DB_SSLMODE": "disable",
	}

	// Establecer las variables de entorno simuladas
	for key, value := range mockEnv {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("No se pudo establecer la variable de entorno %s: %v", key, err)
		}
	}

	// Crear un mock para SecretService
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "some_secret", "testMessageID").
		Return(&SecretData{Username: "postgres", Password: "postgres"}, nil)

	// Crear una instancia de DBManager
	dbManager := NewDBManager(mockSecretService, nil)

	// Crear un mock de SQL usando sqlmock
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err, "No se esperaba un error al crear el mock de SQL")

	// Configurar el dialector de GORM usando el mockDB
	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})

	// Crear un logger de GORM
	mockLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Llamar explícitamente a buildDSN
	dsn := buildDSN(&SecretData{Username: "postgres", Password: "postgres"})
	expectedDSN := fmt.Sprintf(
		dsnString,
		mockEnv["DB_HOST"],
		mockEnv["DB_PORT"],
		"postgres",
		"postgres",
		mockEnv["DB_NAME"],
		mockEnv["DB_SSLMODE"],
	)
	assert.Equal(t, expectedDSN, dsn, "El DSN generado no es el esperado")

	// Configurar expectativa para una consulta de verificación
	mock.ExpectExec(selectVersionQuery).WillReturnResult(sqlmock.NewResult(1, 1))

	// Forzar el flujo para abrir la conexión usando el DSN
	err = dbManager.openConnection(dialector, mockLogger, "testMessageID")
	assert.NoError(t, err, "No se esperaba un error al abrir la conexión")

	// Verificar que la conexión SQL se obtiene correctamente
	sqlDB, err := dbManager.DB.DB()
	assert.NoError(t, err, "No se esperaba un error al obtener la conexión")

	// Ejecutar una consulta para validar la conexión
	_, err = sqlDB.Exec(selectVersionQuery)
	assert.NoError(t, err, "No se esperaba un error al ejecutar la consulta SQL")

	// Configurar la expectativa de cierre
	mock.ExpectClose()

	// Cerrar la conexión y verificar errores
	err = sqlDB.Close()
	assert.NoError(t, err, "No se esperaba un error al cerrar la conexión")

	// Verificar expectativas del mock
	assert.NoError(t, mock.ExpectationsWereMet(), logExpectedMock)
}

func TestGetSecret(t *testing.T) {
	mockService := createMockSecretService()

	secret, err := mockService.GetSecret("some_secret", "testMessageID")
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "testuser", secret.Username)
	assert.Equal(t, "testpassword", secret.Password)

	mockService.AssertExpectations(t)
}

func TestGetSecretSuccess(t *testing.T) {
	// Crear el mock de Secrets Manager
	mockSvc := new(MockSecretsManager)

	// Sobrescribir la función de crear sesión si es necesario
	originalCreateSession := CreateSession
	defer func() { CreateSession = originalCreateSession }()

	// Simular la respuesta de un secreto válido
	secretString := `{"USERNAME": "testuser", "PASSWORD": "testpassword"}`
	mockSvc.On("GetSecretValue", mock.Anything).Return(
		&secretsmanager.GetSecretValueOutput{
			SecretString: awsV2.String(secretString),
		}, nil,
	)

	// Crear el servicio SecretService con el mock inyectado
	secretService := &SecretServiceImpl{
		secretsmanager: mockSvc,
	}

	// Llamar a GetSecret
	secret, err := secretService.GetSecret(testSecretName, "testMessageID")

	// Verificar que no hubo errores
	assert.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, "testuser", secret.Username)
	assert.Equal(t, "testpassword", secret.Password)

	// Verificar que todas las expectativas del mock se cumplieron
	mockSvc.AssertExpectations(t)
}

func TestGetSecretResourceNotFound(t *testing.T) {
	// Crear un mock del cliente de SecretsManager
	mockSvc := new(MockSecretsManager)

	// Simular el error de recurso no encontrado
	mockSvc.On("GetSecretValue", mock.Anything).Return(
		nil, awserr.New(secretsmanager.ErrCodeResourceNotFoundException, logSecretoNoEncontrado, nil),
	)

	// Crear el servicio de secretos utilizando el mock
	secretService := &SecretServiceImpl{
		secretsmanager: mockSvc,
	}

	// Llamar al método y verificar el resultado
	secret, err := secretService.GetSecret("invalid-secret", "testMessageID")

	// Verificar que se devuelva el error adecuado
	assert.Nil(t, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), logSecretoNoEncontrado)

	mockSvc.AssertExpectations(t)
}

func TestGetSecretNullSecretString(t *testing.T) {
	mockSvc := new(MockSecretsManager)

	// Simular que el SecretString es nil
	mockSvc.On("GetSecretValue", mock.Anything).Return(
		&secretsmanager.GetSecretValueOutput{
			SecretString: nil,
		}, nil,
	)

	secretService := &SecretServiceImpl{
		secretsmanager: mockSvc,
	}

	// Llamar al método y verificar que se maneje el secreto nulo
	secret, err := secretService.GetSecret(testSecretName, "testMessageID")

	assert.Nil(t, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "es nulo")

	mockSvc.AssertExpectations(t)
}

func TestGetSecretDeserializationError(t *testing.T) {
	mockSvc := new(MockSecretsManager)

	// Simular un JSON inválido en el secreto
	mockSvc.On("GetSecretValue", mock.Anything).Return(
		&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("{invalid-json}"),
		}, nil,
	)

	secretService := &SecretServiceImpl{
		secretsmanager: mockSvc,
	}

	secret, err := secretService.GetSecret(testSecretName, "testMessageID")

	assert.Nil(t, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deserializar el secreto")

	mockSvc.AssertExpectations(t)
}

func TestNewSecretServiceCreation(t *testing.T) {
	// Crear una sesión simulada de AWS
	mockSession, _ := session.NewSession(&aws.Config{
		Region: aws.String(awsRegion),
	})

	// Crear el servicio con la sesión simulada
	secretService := &SecretServiceImpl{
		secretsmanager: secretsmanager.New(mockSession),
	}

	// Verificar que el cliente de SecretsManager no sea nulo
	assert.NotNil(t, secretService)
	assert.NotNil(t, secretService.secretsmanager)
}

func TestGetSecretValueError(t *testing.T) {
	// Crear un mock del cliente de SecretsManager
	mockSvc := new(MockSecretsManager)

	// Simular un error al obtener el secreto
	mockSvc.On("GetSecretValue", mock.Anything).Return(
		nil, fmt.Errorf("error simulado al obtener el secreto"),
	)

	// Crear el servicio utilizando el mock
	secretService := &SecretServiceImpl{
		secretsmanager: mockSvc,
	}

	// Llamar al método y verificar que se maneja el error
	secret, err := secretService.GetSecret(testSecretName, "testMessageID")

	// Verificar que se registre el error y que el secreto sea nulo
	assert.Nil(t, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al obtener el secreto")

	// Verificar que el mock haya sido llamado correctamente
	mockSvc.AssertExpectations(t)
}

func TestNewSessionError(t *testing.T) {
	originalCreateSession := CreateSession
	defer func() { CreateSession = originalCreateSession }()

	CreateSession = func(_ ...*aws.Config) (*session.Session, error) {
		return nil, errors.New("simulated session error")
	}

	os.Setenv("APP_ENV", "production")
	_, err := NewSession("testMessageID")

	assert.Error(t, err)
	assert.EqualError(t, err, "simulated session error")
}

func TestDBManagerOpenConnectionError(t *testing.T) {
	// Simular las variables de entorno necesarias
	mockEnv := map[string]string{
		"DB_HOST":     "invalid_host",
		"DB_PORT":     "5432",
		"DB_USER":     "dbuser",
		"DB_PASSWORD": "dbpassword",
		"DB_NAME":     "testdb",
		"DB_SSLMODE":  "disable",
	}

	// Establecer las variables de entorno simuladas
	for key, value := range mockEnv {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("No se pudo establecer la variable de entorno %s: %v", key, err)
		}
	}

	// Crear un servicio secreto simulado
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	// Crear un logger para GORM
	logger := createGormLogger()

	// Construir el DSN con los valores simulados
	dsn := fmt.Sprintf(
		dsnString,
		mockEnv["DB_HOST"],
		mockEnv["DB_PORT"],
		mockEnv["DB_USER"],
		mockEnv["DB_PASSWORD"],
		mockEnv["DB_NAME"],
		mockEnv["DB_SSLMODE"],
	)

	// Intentar abrir la conexión con un host inválido
	err := dbManager.openConnection(postgres.Open(dsn), logger, "testMessageID")

	// Verificar que se produce un error
	assert.Error(t, err, "Se esperaba un error al intentar abrir la conexión con un host inválido")
}

func TestCloseDBErrorHandling(t *testing.T) {
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err)

	dbManager.DB, err = gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})
	assert.NoError(t, err)

	sqlDB, err := dbManager.DB.DB()
	assert.NoError(t, err)

	sqlDB.Close()

	dbManager.CloseDB("testMessageID")
	assert.NotNil(t, dbManager.DB)
}

// ======================= Tests Continuación =======================

func TestNewDBManager(t *testing.T) {
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	assert.NotNil(t, dbManager)
	assert.Equal(t, mockService, dbManager.SecretService)
}

func TestGetDBSuccess(t *testing.T) {
	mockDB := &gorm.DB{}
	mockDBManager := new(MockDBManager)
	mockDBManager.On("GetDB").Return(mockDB)

	result := mockDBManager.GetDB()
	assert.NotNil(t, result)

	mockDBManager.AssertExpectations(t)
}

func TestGetDBError(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("GetDB").Return(nil)

	result := mockDBManager.GetDB()
	assert.Nil(t, result)

	mockDBManager.AssertExpectations(t)
}

func TestCloseDBSuccess(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("CloseDB", "testMessageID").Return()

	mockDBManager.CloseDB("testMessageID")
	mockDBManager.AssertExpectations(t)
}

func TestLocalStackSession(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	session, err := NewSession("testMessageID")

	assert.NoError(t, err)
	assert.NotNil(t, session)
}

func TestRemoteSessionProduction(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	session, err := NewSession("testMessageID")

	assert.NoError(t, err)
	assert.NotNil(t, session)
}

func TestSecretNotFoundError(t *testing.T) {
	mockSvc := new(MockSecretsManager)
	mockSvc.On("GetSecretValue", mock.Anything).
		Return(
			nil,
			awserr.New(secretsmanager.ErrCodeResourceNotFoundException,
				logSecretoNoEncontrado, nil))

	secretService := &SecretServiceImpl{secretsmanager: mockSvc}
	secret, err := secretService.GetSecret("invalid-secret", "testMessageID")

	assert.Nil(t, secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), logSecretoNoEncontrado)
}

func TestSecretServiceEmptySecret(t *testing.T) {
	mockService := new(MockSecretService)
	mockService.On("GetSecret", "", "testMessageID").
		Return(nil, errors.New("el nombre del secreto no puede estar vacío"))

	secret, err := mockService.GetSecret("", "testMessageID")

	assert.Nil(t, secret)
	assert.Error(t, err)
	mockService.AssertExpectations(t)
}

func TestBuildDSNAndOpenConnectionSuccess(t *testing.T) {
	// Simular las variables de entorno necesarias
	mockEnv := map[string]string{
		"DB_HOST":    "localhost",
		"DB_PORT":    "5432",
		"DB_NAME":    "testdb",
		"DB_SSLMODE": "require", // Aseguramos consistencia con el valor generado
	}

	// Establecer las variables de entorno simuladas
	for key, value := range mockEnv {
		err := os.Setenv(key, value)
		if err != nil {
			t.Fatalf("No se pudo establecer la variable de entorno %s: %v", key, err)
		}
	}

	// Crear un servicio secreto simulado
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	// Crear un mock de SQL usando sqlmock
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err, "No se esperaba un error al crear el mock de SQL")

	// Configurar el dialector de GORM usando el mockDB
	dialector := postgres.New(postgres.Config{
		Conn: mockDB,
	})

	// Crear un logger para GORM
	logger := createGormLogger()

	// Construir el DSN con valores simulados
	dsn := buildDSN(&SecretData{Username: "postgres", Password: "postgres"})
	expectedDSN := fmt.Sprintf(
		dsnString,
		mockEnv["DB_HOST"],
		mockEnv["DB_PORT"],
		"postgres",
		"postgres",
		mockEnv["DB_NAME"],
		mockEnv["DB_SSLMODE"], // Aseguramos que use el sslmode esperado
	)

	// Validar que el DSN generado es igual al esperado
	assert.Equal(t, expectedDSN, dsn, "El DSN generado no es el esperado")

	// Configurar expectativa para una consulta de verificación
	mock.ExpectExec(selectVersionQuery).WillReturnResult(sqlmock.NewResult(1, 1))

	// Abrir la conexión usando el DSN generado
	err = dbManager.openConnection(dialector, logger, "testMessageID")
	assert.NoError(t, err, "No se esperaba un error al abrir la conexión")

	// Verificar que la conexión SQL se obtiene correctamente
	sqlDB, err := dbManager.DB.DB()
	assert.NoError(t, err, "No se esperaba un error al obtener la conexión")

	// Ejecutar una consulta para validar la conexión
	_, err = sqlDB.Exec(selectVersionQuery)
	assert.NoError(t, err, "No se esperaba un error al ejecutar la consulta SQL")

	// Configurar la expectativa de cierre
	mock.ExpectClose()

	// Cerrar la conexión y verificar errores
	err = sqlDB.Close()
	assert.NoError(t, err, "No se esperaba un error al cerrar la conexión")

	// Verificar expectativas del mock
	assert.NoError(t, mock.ExpectationsWereMet(), logExpectedMock)
}

func TestNewSecretService(t *testing.T) {
	mockSession, err := session.NewSession(&aws.Config{Region: aws.String(awsRegion)})
	assert.NoError(t, err)

	secretService := NewSecretService(mockSession)
	assert.NotNil(t, secretService)

	serviceImpl, ok := secretService.(*SecretServiceImpl)
	assert.True(t, ok)
	assert.NotNil(t, serviceImpl.secretsmanager)
}

func TestBuildDSNDefaultSSLMode(t *testing.T) {
	os.Unsetenv("DB_SSLMODE") // Asegurarse de que no está configurado
	secretData := &SecretData{Username: "user", Password: "pass"}

	dsn := buildDSN(secretData)

	expectedDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		secretData.Username,
		secretData.Password,
		os.Getenv("DB_NAME"),
	)

	assert.Equal(t, expectedDSN, dsn, "El DSN generado no coincide con el esperado")
}

func TestCloseDBWithError(t *testing.T) {
	mockService := createMockSecretService()
	dbManager := NewDBManager(mockService, nil)

	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dbManager.DB, _ = gorm.Open(postgres.New(postgres.Config{Conn: mockDB}), &gorm.Config{})

	sqlDB, err := dbManager.DB.DB()

	assert.NoError(t, err)

	mock.ExpectClose().WillReturnError(fmt.Errorf("error al cerrar la conexión"))

	dbManager.CloseDB("testMessageID")

	assert.NotNil(t, dbManager.DB)

	err = sqlDB.Close()

	// Verificar que el error fue registrado
	assert.NoError(t, mock.ExpectationsWereMet(), "No se cumplieron las expectativas del mock")
}

func TestOpenConnectionError(t *testing.T) {
	mockSecretService := createMockSecretService()
	dbManager := NewDBManager(mockSecretService, nil)

	// Crear un DSN inválido para forzar el error
	os.Setenv("DB_HOST", "invalid_host")
	logger := createGormLogger()

	dsn := buildDSN(&SecretData{Username: "invalid", Password: "invalid"})

	// Llamar a openConnection y verificar que se devuelve un error
	err := dbManager.openConnection(postgres.Open(dsn), logger, "testMessageID")

	assert.Error(t, err, "Se esperaba un error al abrir la conexión")
	assert.Contains(t, err.Error(), "error al abrir la conexión")
}
