package email

import (
	"errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/connection"
	"net/smtp"
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

// Mock de smtpSendMailFunc para simular el envío de correo sin realizar la operación real.
func mockSendMailSuccess(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return nil // Simula éxito
}

func mockSendMailError(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return errors.New("error enviando el correo") // Simula un error
}

// Mock para forzar un retraso, simulando un timeout.
func mockSendMailTimeout(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	time.Sleep(100 * time.Millisecond) // Fuerza un retraso mayor al timeout del servicio
	return nil
}
func TestNewSMTPEmailService_DefaultTimeout(t *testing.T) {
	mockSecretService := new(MockSecretService)
	secretName := "my-smtp-secrets"
	secretData := &connection.SecretData{
		Username: "user",
		Password: "pass",
	}

	// Simular la obtención exitosa del secreto
	mockSecretService.On("GetSecret", secretName, "test-message-id").Return(secretData, nil)

	// Simular las variables de entorno sin definir el timeout
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")

	// Crear el servicio
	service, err := NewSMTPEmailService(mockSecretService, "test-message-id")

	// Verificar que no haya error y que el timeout sea el valor por defecto (15 segundos)
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, 15*time.Second, service.timeout)

	mockSecretService.AssertExpectations(t)
}

// Test que verifica el envío exitoso de un correo
func TestSMTPEmailService_SendEmail_Success(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		sendMail: mockSendMailSuccess,
		timeout:  10 * time.Second,
	}

	err := service.SendEmail("sender@test.com", "recipient@test.com", "Test Subject", "Test Body", "test-message-id")
	assert.NoError(t, err)
}

// Test que verifica el manejo de timeout
func TestSMTPEmailService_SendEmail_Timeout(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		sendMail: mockSendMailTimeout,  // Simular un retraso largo
		timeout:  1 * time.Millisecond, // Timeout muy corto para forzar el error
	}

	err := service.SendEmail("sender@test.com", "recipient@test.com", "Test Subject", "Test Body", "test-message-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// Test que verifica el manejo de error durante el envío del correo
func TestSMTPEmailService_SendEmail_Error(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		sendMail: mockSendMailError, // Simular error
		timeout:  10 * time.Second,
	}

	err := service.SendEmail("sender@test.com", "recipient@test.com", "Test Subject", "Test Body", "test-message-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error enviando el correo")
}

// Test que verifica la configuración incompleta del servicio SMTP
func TestSMTPEmailService_SendEmail_IncompleteConfig(t *testing.T) {
	service := &SMTPEmailService{
		server:   "",
		port:     "",
		username: "",
		password: "",
		sendMail: mockSendMailSuccess,
		timeout:  10 * time.Second,
	}

	err := service.SendEmail("sender@test.com", "recipient@test.com", "Test Subject", "Test Body", "test-message-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuración SMTP incompleta")
}

// Test que verifica el caso donde los destinatarios están mal formados o son inválidos.
func TestSMTPEmailService_SendEmail_InvalidRecipient(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		sendMail: mockSendMailSuccess,
		timeout:  10 * time.Second,
	}

	// Caso de destinatarios vacíos
	err := service.SendEmail(
		"sender@test.com", // Remitente
		"",                // Destinatarios vacíos
		"Test Subject",    // Asunto
		"Test Body",       // Cuerpo
		"test-message-id", // Message ID
	)

	// Verificamos que se produzca un error por destinatarios inválidos
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error: no se especificaron destinatarios")
}

func TestNewSMTPEmailService_Success(t *testing.T) {
	mockSecretService := new(MockSecretService)
	secretName := "my-smtp-secrets"
	secretData := &connection.SecretData{
		Username: "user",
		Password: "pass",
	}

	// Simular la obtención exitosa del secreto
	mockSecretService.On("GetSecret", secretName, "test-message-id").Return(secretData, nil)

	// Simular las variables de entorno
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")

	// Simular el valor de timeout usando Viper
	viper.Set("SMTP_TIMEOUT", "10")

	// Crear el servicio
	service, err := NewSMTPEmailService(mockSecretService, "test-message-id")

	// Verificar que no haya error y que los valores sean correctos
	assert.NoError(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "smtp.test.com", service.server)
	assert.Equal(t, "587", service.port)
	assert.Equal(t, "user", service.username)
	assert.Equal(t, "pass", service.password)
	assert.Equal(t, 10*time.Second, service.timeout)

	mockSecretService.AssertExpectations(t)
}

func TestNewSMTPEmailService_ErrorGettingSecret(t *testing.T) {
	mockSecretService := new(MockSecretService)
	secretName := "my-smtp-secrets"

	// Simular que la obtención del secreto devuelve un error
	mockSecretService.On("GetSecret", secretName, "test-message-id").Return((*connection.SecretData)(nil), errors.New("error obteniendo secreto"))

	// Simular las variables de entorno
	t.Setenv("SECRETS_SMTP", secretName)
	t.Setenv("SMTP_SERVER", "smtp.test.com")
	t.Setenv("SMTP_PORT", "587")

	// Crear el servicio (debe fallar)
	service, err := NewSMTPEmailService(mockSecretService, "test-message-id")

	// Verificar que haya un error y que el servicio sea nil
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "error obteniendo secreto")

	mockSecretService.AssertExpectations(t)
}

func TestSMTPEmailService_SendEmail_ErrorConvertingRecipients(t *testing.T) {
	service := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user",
		password: "pass",
		sendMail: mockSendMailSuccess, // No importa el resultado del envío en este caso
		timeout:  10 * time.Second,
	}

	// Probar un destinatario mal formado
	err := service.SendEmail("sender@test.com", string([]byte{0x7f}), "Test Subject", "Test Body", "test-message-id")

	// Verificar que se retorne el error esperado
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al convertir destinatarios a JSON")
}
