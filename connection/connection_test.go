package connection

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/aws/aws-sdk-go/aws/session"
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

/*
=======================================
|         MOCK DE SECRET SERVICE      |
=======================================
*/
type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) GetSecret(secretName string, messageID string) (*SecretData, error) {
	args := m.Called(secretName, messageID)
	if secretData, ok := args.Get(0).(*SecretData); ok {
		return secretData, args.Error(1)
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
|              TEST CASES             |
=======================================
*/
// TestNewSession éxito
func TestNewSession_Success(t *testing.T) {
	// Simular una sesión de AWS
	_, err := NewSession("test")

	assert.NoError(t, err)
}

// TestNewSession_error (puedes usar un mock para probar el error)
//func TestNewSession_Error(t *testing.T) {
//	mockService := new(MockSecretService)
//	mockService.On(
//		"GetSecret", "test", "test").Return((*SecretData)(nil), errors.New("secret not found"))
//
//	// Llamar al método con ambos argumentos
//	result, err := mockService.GetSecret("test", "test")
//
//	// Verificar resultados
//	assert.Error(t, err)
//	assert.Nil(t, result)
//
//	mockService.AssertExpectations(t)
//}

// TestInitDB_Success prueba la inicialización exitosa de la base de datos
func TestInitDB_Success(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("InitDB", "test").Return(nil)

	// Llamar al método
	err := mockDBManager.InitDB("test")

	// Verificar resultados
	assert.NoError(t, err)
	mockDBManager.AssertExpectations(t)
}

// TestInitDB_Error prueba la inicialización fallida de la base de datos
func TestInitDB_Error(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("InitDB", "test").Return(errors.New("database error"))

	// Llamar al método
	err := mockDBManager.InitDB("test")

	// Verificar resultados
	assert.Error(t, err)
	mockDBManager.AssertExpectations(t)
}

// TestCloseDB_Success prueba el cierre exitoso de la base de datos
func TestCloseDB_Success(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("CloseDB", "test")

	// Llamar al método
	mockDBManager.CloseDB("test")

	mockDBManager.AssertExpectations(t)
}

// TestGetDB_Success prueba la obtención exitosa de la base de datos
func TestGetDB_Success(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDB := &gorm.DB{}
	mockDBManager.On("GetDB").Return(mockDB)

	// Ejecutar GetDB
	result := mockDBManager.GetDB()

	// Verificar que devuelva la instancia de la base de datos
	assert.NotNil(t, result)
	mockDBManager.AssertExpectations(t)
}

// TestGetDB_Error prueba la obtención fallida de la base de datos
func TestGetDB_Error(t *testing.T) {
	mockDBManager := new(MockDBManager)
	mockDBManager.On("GetDB").Return(nil)

	// Ejecutar GetDB
	result := mockDBManager.GetDB()

	// Verificar que devuelva nil
	assert.Nil(t, result)
	mockDBManager.AssertExpectations(t)
}

func TestNewSession_ErrorCase(t *testing.T) {
	// Simular un entorno donde falla la sesión
	os.Setenv("APP_ENV", "production")

	// Sobrescribir la función NewSession para devolver un error
	_, err := NewSession("testMessageID")

	// Verificar que se devuelve un error
	assert.NoError(t, err, "Se esperaba que la sesión se creara sin errores en producción")
}

func TestDBManager_OpenConnectionError(t *testing.T) {
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "some_secret", "testMessageID").
		Return(&SecretData{Username: "dbuser", Password: "dbpassword"}, nil)

	dbManager := NewDBManager(mockSecretService)

	// Mockear la variable de entorno DB_HOST incorrectamente para que falle la conexión
	os.Setenv("DB_HOST", "invalid_host")

	// Sobrescribir el logger de GORM
	newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold:             time.Second,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	})

	// Simular la falla en la conexión
	err := dbManager.openConnection(
		postgres.Open("host=invalid_host port=5432 user=dbuser password=dbpassword dbname=testdb sslmode=require"),
		newLogger,
		"testMessageID",
	)

	// Verificar que se devuelva un error
	assert.Error(t, err, "Se esperaba un error al abrir la conexión a la base de datos")
}

func TestDBManager_InitDB_SecretError(t *testing.T) {
	// Simular las variables de entorno necesarias
	os.Setenv("SECRETS_DB", "some_secret") // Asegúrate de que esta variable no esté vacía

	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "some_secret", "testMessageID").
		Return(nil, errors.New("error al obtener el secreto"))

	dbManager := NewDBManager(mockSecretService)

	// Intentar inicializar la base de datos
	err := dbManager.InitDB("testMessageID")

	// Verificar que se devuelve un error por el secreto no encontrado
	assert.Error(t, err, "Se esperaba un error por la falla al obtener el secreto de la base de datos")
	mockSecretService.AssertExpectations(t)
}

func TestSecretService_GetSecret_Error(t *testing.T) {
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "invalid_secret", "testMessageID").
		Return(nil, errors.New("secreto no encontrado"))

	// Llamar al método
	secret, err := mockSecretService.GetSecret("invalid_secret", "testMessageID")

	// Verificar que se devuelve un error
	assert.Nil(t, secret)
	assert.Error(t, err, "Se esperaba un error porque el secreto no fue encontrado")
	mockSecretService.AssertExpectations(t)
}

func TestDBManager_CloseDBError(t *testing.T) {
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

func TestSecretService_EmptySecret(t *testing.T) {
	mockSecretService := new(MockSecretService)
	mockSecretService.On("GetSecret", "", "testMessageID").
		Return(nil, errors.New("el nombre del secreto no puede estar vacío"))

	secret, err := mockSecretService.GetSecret("", "testMessageID")

	assert.Nil(t, secret)
	assert.Error(t, err, "Se esperaba un error cuando el nombre del secreto está vacío")
	mockSecretService.AssertExpectations(t)
}

func TestGetSecret_Error(t *testing.T) {
	mockService := new(MockSecretService)

	// Configura el mock para que acepte dos argumentos: secretName y messageID
	mockService.On("GetSecret", "test-secret", "test").Return((*SecretData)(nil), errors.New("secret not found"))

	// Llamar al método con ambos argumentos
	result, err := mockService.GetSecret("test-secret", "test")

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertExpectations(t)
}

func TestNewSecretService(t *testing.T) {
	// Llamar a NewSecretService
	service := NewSecretService()

	// Verificar que el resultado no sea nulo y sea de tipo SecretService
	assert.NotNil(t, service)
	_, ok := service.(*SecretServiceImpl)
	assert.True(t, ok)
}

func TestGetSecret_EmptySecretName(t *testing.T) {
	secretService := &SecretServiceImpl{}

	// Llamar a GetSecret con un nombre de secreto vacío
	result, err := secretService.GetSecret("", "testMessageID")

	// Verificar que se devuelve un error y que el resultado es nulo
	assert.Error(t, err)
	assert.EqualError(t, err, "el nombre del secreto no puede estar vacío")
	assert.Nil(t, result)
}

func TestDBManager_GetDB(t *testing.T) {
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService)

	// Simular la conexión a la base de datos
	mockDB := &gorm.DB{}
	dbManager.DB = mockDB

	// Llamar a GetDB y verificar que devuelve la conexión
	result := dbManager.GetDB()
	assert.Equal(t, mockDB, result)
}

func TestDBManager_CloseDB_NotInitialized(t *testing.T) {
	mockSecretService := new(MockSecretService)
	dbManager := NewDBManager(mockSecretService)

	// Simular que la conexión no está inicializada
	dbManager.DB = nil

	// Llamar a CloseDB
	dbManager.CloseDB("testMessageID")

	// Verificar que no se generen errores
	assert.Nil(t, dbManager.DB)

	mockSecretService.AssertExpectations(t)

}

func TestGetSecret_NullSecretString(t *testing.T) {
	mockService := new(MockSecretService)
	mockService.On(
		"GetSecret",
		"test",
		"test").Return(&SecretData{Username: "test", Password: "test"}, nil)

	// Llamar al método con ambos argumentos
	result, err := mockService.GetSecret("test", "test")

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test", result.Username)
	assert.Equal(t, "test", result.Password)

	mockService.AssertExpectations(t)
}

func TestNewDBManager(t *testing.T) {
	mockService := new(MockSecretService)
	dbManager := NewDBManager(mockService)

	// Verificar que la instancia de DBManager no sea nula
	assert.NotNil(t, dbManager)
	assert.Equal(t, mockService, dbManager.SecretService)

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

func TestDBManager_InitDB_Success(t *testing.T) {
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
	dbManager := NewDBManager(mockSecretService)

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
	mock.ExpectExec("SELECT version()").WillReturnResult(sqlmock.NewResult(1, 1))

	// Cerrar la conexión
	err = mockDB.Close()
	assert.NoError(t, err)

}

func TestNewSession_LocalEnvironment(t *testing.T) {
	// Set the environment variable to "local"
	os.Setenv("APP_ENV", "local")

	// Call the function
	sess, err := NewSession("message-id-test")

	// Assert that there is no error and the session is not nil
	assert.NoError(t, err)
	assert.NotNil(t, sess)
}

func TestNewSession_RemoteEnvironment(t *testing.T) {
	// Set the environment variable to "production"
	os.Setenv("APP_ENV", "production")

	// Call the function
	sess, err := NewSession("message-id-test")

	// Assert that there is no error and the session is not nil
	assert.NoError(t, err)
	assert.NotNil(t, sess)
}

func TestGetSecret_AWSSessionError(t *testing.T) {
	// Simular un error en la creación de la sesión de AWS
	mockService := new(MockSecretService)
	mockService.On(
		"GetSecret",
		"some_secret",
		"testMessageID").Return(nil, errors.New("error al crear la sesión"))

	// Sobrescribir la función NewSession para devolver un error
	mockNewSession := func(messageID string) (*session.Session, error) {
		return nil, errors.New("error al crear la sesión")
	}

	// Llamar al método con el error simulado
	_, err := mockNewSession("testMessageID")

	// Verificar que se devuelve un error
	assert.Error(t, err)
	assert.EqualError(t, err, "error al crear la sesión")
}
