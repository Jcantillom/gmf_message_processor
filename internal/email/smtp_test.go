package email

import (
	"bytes"
	"errors"
	"log"
	"net/smtp"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Función de ayuda para capturar la salida del log
func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	f()
	return buf.String()
}

func TestSendEmail_Success(t *testing.T) {
	// Mock de las variables de entorno necesarias
	os.Setenv("SMTP_SERVER", "smtp.example.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "user@example.com")
	os.Setenv("SMTP_PASSWORD", "password")

	// Mock de la función de envío de correo
	smtpSendMailWrapper = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}

	// Redirigir la salida de logs y ejecutar la función de envío de correo
	output := captureOutput(func() {
		err := SendEmail("from@example.com", "to1@example.com,to2@example.com", "Asunto de Prueba", "Cuerpo del correo")
		assert.NoError(t, err, "No debería haber error al enviar el correo")
	})

	// Verificar que el log contenga los destinatarios
	assert.Contains(t, output, "Enviando correo a los siguientes destinatarios 📧")
	assert.Contains(t, output, "Correo electrónico enviado con éxito a 📤")
}

func TestSendEmail_MissingSMTPConfig(t *testing.T) {
	// Mock de las variables de entorno necesarias
	os.Setenv("SMTP_SERVER", "")
	os.Setenv("SMTP_PORT", "")
	os.Setenv("SMTP_USER", "")
	os.Setenv("SMTP_PASSWORD", "")

	// Mock de la función de envío de correo (no se necesita en esta prueba, pero se puede añadir por consistencia)
	smtpSendMailWrapper = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return nil
	}

	// Redirigir la salida de logs y ejecutar la función de envío de correo
	output := captureOutput(func() {
		err := SendEmail("from@example.com", "to@example.com", "Asunto de Prueba", "Cuerpo del correo")
		assert.Error(t, err, "Debería haber un error debido a la configuración SMTP incompleta")
		assert.Contains(t, err.Error(), "configuración SMTP incompleta en las variables de entorno")
	})

	// Verificar que el log no contenga los destinatarios porque falló antes
	assert.NotContains(t, output, "Enviando correo a los siguientes destinatarios 📧")
}

func TestSendEmail_ErrorInSMTP(t *testing.T) {
	// Mock de las variables de entorno necesarias
	os.Setenv("SMTP_SERVER", "invalid.server.com")
	os.Setenv("SMTP_PORT", "587")
	os.Setenv("SMTP_USER", "user@example.com")
	os.Setenv("SMTP_PASSWORD", "password")

	// Mock de la función de envío de correo para simular un error
	smtpSendMailWrapper = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("error enviando el correo electrónico")
	}

	// Ejecutar la función de envío de correo
	err := SendEmail(
		"from@example.com",
		"to1@example.com,to2@example.com",
		"Asunto de Prueba",
		"Cuerpo del correo",
	)

	// Verificar que haya un error y que sea el esperado
	assert.Error(t, err, "Debería haber un error al enviar el correo")
	assert.Contains(t, err.Error(), "error enviando el correo electrónico", "El error debe contener 'error enviando el correo electrónico'")
}
