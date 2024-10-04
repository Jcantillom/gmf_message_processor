package email

import (
	"errors"
	"net/smtp"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock para la función smtpSendMail
type MockSMTP struct {
	mock.Mock
}

func (m *MockSMTP) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	args := m.Called(addr, a, from, to, msg)
	return args.Error(0)
}

func TestSendEmail_IncompleteSMTPConfig(t *testing.T) {
	// Guarda las configuraciones SMTP actuales para restaurarlas al final de la prueba
	originalServer := os.Getenv("SMTP_SERVER")
	originalPort := os.Getenv("SMTP_PORT")
	originalUser := os.Getenv("SMTP_USER")
	originalPassword := os.Getenv("SMTP_PASSWORD")

	// Limpia las configuraciones SMTP
	os.Setenv("SMTP_SERVER", "")
	os.Setenv("SMTP_PORT", "")
	os.Setenv("SMTP_USER", "")
	os.Setenv("SMTP_PASSWORD", "")

	// Crear instancia del servicio SMTP
	emailService := NewSMTPEmailService()

	// Realizar la prueba
	err := emailService.SendEmail("from@test.com", "to@test.com", "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Equal(t, "error: configuración SMTP incompleta en las variables de entorno", err.Error())

	// Restaurar las configuraciones originales
	os.Setenv("SMTP_SERVER", originalServer)
	os.Setenv("SMTP_PORT", originalPort)
	os.Setenv("SMTP_USER", originalUser)
	os.Setenv("SMTP_PASSWORD", originalPassword)
}
func TestSendEmail_SendMailError(t *testing.T) {
	mockSMTP := new(MockSMTP)
	mockSMTP.On(
		"SendMail",
		"smtp.test.com:587",
		mock.Anything,
		"from@test.com",
		[]string{"to@test.com"},
		mock.Anything,
	).Return(errors.New("error enviando el correo electrónico"))

	emailService := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user@test.com",
		password: "password",
		sendMail: mockSMTP.SendMail,
	}

	// Realizar la prueba
	err := emailService.SendEmail("from@test.com", "to@test.com", "Test Subject", "Test Body")
	assert.Error(t, err)
	assert.Equal(t, "error enviando el correo electrónico: error enviando el correo electrónico", err.Error())

	mockSMTP.AssertExpectations(t)
}
func TestSendEmail_Success(t *testing.T) {
	mockSMTP := new(MockSMTP)
	mockSMTP.On(
		"SendMail",
		"smtp.test.com:587",
		mock.Anything,
		"from@test.com",
		[]string{"to@test.com"},
		mock.Anything,
	).Return(nil)

	emailService := &SMTPEmailService{
		server:   "smtp.test.com",
		port:     "587",
		username: "user@test.com",
		password: "password",
		sendMail: mockSMTP.SendMail,
	}

	// Realizar la prueba
	err := emailService.SendEmail("from@test.com", "to@test.com", "Test Subject", "Test Body")
	assert.NoError(t, err)

	mockSMTP.AssertExpectations(t)
}
func TestSendEmail_JSONConversionError(t *testing.T) {
	mockSMTP := new(MockSMTP)

	// Simular configuración SMTP correcta
	os.Setenv("SMTP_SERVER", "smtp.test.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "user@test.com")
	os.Setenv("SMTP_PASSWORD", "password")

	emailService := &SMTPEmailService{
		server:   os.Getenv("SMTP_SERVER"),
		port:     os.Getenv("SMTP_PORT"),
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		sendMail: mockSMTP.SendMail,
	}

	// Configurar mock para que no se espere la llamada a SendMail
	mockSMTP.On(
		"SendMail",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil).Maybe()

	// Usar una lista de destinatarios que cause un error JSON
	invalidRecipient := "\x7f"
	err := emailService.SendEmail(
		"from@test.com",
		invalidRecipient,
		"Test Subject",
		"Test Body")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al convertir destinatarios a JSON")

	// Verificar que SendMail nunca fue llamado
	mockSMTP.AssertNotCalled(t, "SendMail")
}
