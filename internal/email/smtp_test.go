package email

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/connection"
	"gopkg.in/gomail.v2"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock del servicio de secretos para simular la obtención de secretos.
type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) GetSecret(secretName string, messageID string) (*connection.SecretData, error) {
	args := m.Called(secretName, messageID)
	return args.Get(0).(*connection.SecretData), args.Error(1) // Retorna un puntero a SecretData
}

type MockLogger struct {
	loggedMessage string
}

func (m *MockLogger) LogError(message string, err error, messageID string) {
	m.loggedMessage = fmt.Sprintf("%s - %v", message, err)
}

func (m *MockLogger) LogInfo(message, messageID string) {
	m.loggedMessage = message
}

type MockDialer struct{}

// Simula el envío de correo sin error.
func (d *MockDialer) DialAndSend(m ...*gomail.Message) error {
	return nil
}

// Simula un error al enviar el correo.
type ErrorMockDialer struct{}

func (d *ErrorMockDialer) DialAndSend(m ...*gomail.Message) error {
	return fmt.Errorf("error enviando el correo")
}

func mockStat(filename string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

const secretName = "my-smtp-secrets"
const testMessageID = "test-message-id"
const smtpServerTest = "smtp.test.com"
const senderEmailTest = "sender@test.com"
const recipientEmailTest = "recipient@test.com"
const testSubject = "Test Subject"
const testBody = "Test Body"
const testPathImage = "../../images/Casitadavivienda.png"
const testImageName = "logo.png"

func TestNewSMTPEmailServiceDefaultTimeout(t *testing.T) {
	mockSecretService := new(MockSecretService)

	secretData := &connection.SecretData{
		Username: "user",
		Password: "pass",
	}

	// Simular la obtención exitosa del secreto
	mockSecretService.On("GetSecret", secretName, testMessageID).Return(secretData, nil)

	// Simular las variables de entorno sin definir el timeout
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", smtpServerTest)
	t.Setenv("SMTP_PORT", "587")

	// Crear el servicio
	service, err := NewSMTPEmailService(mockSecretService, testMessageID, &MockLogger{})

	// Verificar que no haya error y que el timeout sea el valor por defecto (15 segundos)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, 15*time.Second, service.timeout)

	mockSecretService.AssertExpectations(t)
}

// Test que verifica el envío exitoso de un correo
func TestSMTPEmailServiceSendEmailSuccess(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   &MockDialer{}, // Usar el mock del dialer
		timeout:  10 * time.Second,
	}

	err := service.SendEmail(
		"sender@test.com", "recipient@test.com", "Test Subject", "Test Body",
		"../../images/Casitadavivienda.png", "logo.png", "test-message-id",
	)

	assert.NoError(t, err)
}

// Test que verifica el manejo de timeout
func TestSMTPEmailServiceSendEmailTimeout(t *testing.T) {
	mockDialer := gomail.NewDialer("smtp.test.com", 587, "user", "pass")
	service := &SMTPEmailService{
		server:   smtpServerTest,
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   mockDialer,
		timeout:  1 * time.Millisecond, // Timeout muy corto para forzar el error
	}

	err := service.SendEmail(
		senderEmailTest,
		recipientEmailTest,
		testSubject,
		testBody,
		testPathImage,
		testImageName,
		testMessageID,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// Test que verifica el manejo de error durante el envío del correo
func TestSMTPEmailServiceSendEmailError(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   &ErrorMockDialer{}, // Usar el mock que genera error
		timeout:  10 * time.Second,
	}

	err := service.SendEmail(
		"sender@test.com", "recipient@test.com", "Test Subject", "Test Body",
		"../../images/Casitadavivienda.png", "logo.png", "test-message-id",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error enviando el correo")
}

// Test que verifica la configuración incompleta del servicio SMTP
func TestSMTPEmailServiceSendEmailIncompleteConfig(t *testing.T) {
	mockDialer := gomail.NewDialer("smtp.test.com", 587, "user", "pass")
	service := &SMTPEmailService{
		server:   "",
		port:     "",
		username: "",
		password: "",
		dialer:   mockDialer,
		timeout:  10 * time.Second,
	}

	err := service.SendEmail(
		senderEmailTest,
		recipientEmailTest,
		testSubject,
		testBody,
		testPathImage,
		testImageName,
		testMessageID,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuración SMTP incompleta")
}

// Test que verifica el caso donde los destinatarios están mal formados o son inválidos.
func TestSMTPEmailServiceSendEmailInvalidRecipient(t *testing.T) {
	service := &SMTPEmailService{
		server:   smtpServerTest,
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   &MockDialer{}, // Usar el mock del dialer
		timeout:  10 * time.Second,
	}

	// Caso de destinatarios vacíos
	err := service.SendEmail(
		senderEmailTest,
		"", // Destinatarios vacíos
		testSubject,
		testBody,
		testPathImage,
		testImageName,
		testMessageID,
	)

	// Verificar que se genere el error esperado
	// Verificar que se genere el error esperado
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "error: no se especificaron destinatarios")
	} else {
		t.Fatalf("Se esperaba un error, pero no se recibió ninguno")
	}
}

func TestNewSMTPEmailServiceSuccess(t *testing.T) {
	mockSecretService := new(MockSecretService)
	secretData := &connection.SecretData{
		Username: "user",
		Password: "pass",
	}

	// Simular la obtención exitosa del secreto
	mockSecretService.On(
		"GetSecret", secretName, testMessageID).Return(secretData, nil)

	// Simular las variables de entorno
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", smtpServerTest)
	t.Setenv("SMTP_PORT", "587")

	// Simular el valor de timeout usando Viper
	viper.Set("SMTP_TIMEOUT", "10")

	// Crear el servicio
	service, err := NewSMTPEmailService(mockSecretService, testMessageID, &MockLogger{})

	// Verificar que no haya error y que los valores sean correctos
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, smtpServerTest, service.server)
	assert.Equal(t, "587", service.port)
	assert.Equal(t, "user", service.username)
	assert.Equal(t, "pass", service.password)
	assert.Equal(t, 10*time.Second, service.timeout)

	mockSecretService.AssertExpectations(t)
}

func TestNewSMTPEmailServiceErrorGettingSecret(t *testing.T) {
	mockSecretService := new(MockSecretService)

	// Simular que la obtención del secreto devuelve un error
	mockSecretService.On(
		"GetSecret",
		secretName,
		testMessageID).
		Return((*connection.SecretData)(nil), errors.New("error obteniendo secreto"))

	// Simular las variables de entorno
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", smtpServerTest)
	t.Setenv("SMTP_PORT", "587")

	// Crear el servicio (debe fallar)
	service, err := NewSMTPEmailService(mockSecretService, testMessageID, &MockLogger{})

	// Verificar que haya un error y que el servicio sea nil
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "error obteniendo secreto")

	mockSecretService.AssertExpectations(t)
}

func TestSMTPEmailServiceSendEmailErrorConvertingRecipients(t *testing.T) {
	mockDialer := gomail.NewDialer("smtp.test.com", 587, "user", "pass")
	service := &SMTPEmailService{
		server:   smtpServerTest,
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   mockDialer,
		timeout:  10 * time.Second,
	}

	// Probar un destinatario mal formado
	err := service.SendEmail(
		senderEmailTest,
		string([]byte{0x7f}),
		testSubject,
		testBody,
		testPathImage,
		testImageName,
		testMessageID,
	)

	// Verificar que se retorne el error esperado
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al convertir destinatarios a JSON")
}

func TestSendEmailImageNotFound(t *testing.T) {
	// Crear un mock del logger.
	mockLogger := &MockLogger{}

	// Crear el servicio de correo con el mock del logger.
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		dialer:   &MockDialer{},
		timeout:  10 * time.Second,
		logger:   mockLogger,
	}

	// Intentar enviar el correo con una imagen inexistente.
	err := service.SendEmail(
		"sender@test.com", "recipient@test.com", "Test Subject", "Test Body",
		"/path/to/nonexistent-image.png", "logo.png", "test-message-id",
	)

	// Verificar que se produjo el error esperado.
	assert.Error(t, err)

	// Verificar que el mensaje registrado por el logger sea el esperado.
	expectedMessage := "La imagen /path/to/nonexistent-image.png no existe en la ruta proporcionada"
	assert.Contains(t, mockLogger.loggedMessage, expectedMessage)

	// Verificar que el error contenga "no such file or directory".
	assert.Contains(t, mockLogger.loggedMessage, "no such file or directory")
}
