package email

import (
	"errors"
	"net/smtp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
