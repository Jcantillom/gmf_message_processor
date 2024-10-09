package email

import (
	"context"
	"fmt"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/logs"
	"net/smtp"
	"os"
	"strings"
)

var ctx = context.TODO()

// SMTPEmailService implementa EmailService utilizando SMTP.
type SMTPEmailService struct {
	server   string
	port     string
	username string
	password string
	sendMail smtpSendMailFunc
}

// smtpSendMailFunc es una función de envío de correo electrónico SMTP.
type smtpSendMailFunc func(
	addr string, a smtp.Auth, from string, to []string, msg []byte) error

// NewSMTPEmailService crea una nueva instancia de SMTPEmailService usando SecretService para obtener las credenciales SMTP.
func NewSMTPEmailService(secretService connection.SecretService) (*SMTPEmailService, error) {
	secretName := os.Getenv("SECRETS_SMTP")
	secretData, err := secretService.GetSecret(secretName)
	if err != nil {
		logs.LogError("Error al obtener las credenciales SMTP desde Secrets Manager: %v", err)
		return nil, err
	}

	return &SMTPEmailService{
		server:   os.Getenv("SMTP_SERVER"),
		port:     os.Getenv("SMTP_PORT"),
		username: secretData.Username,
		password: secretData.Password,
		sendMail: smtp.SendMail,
	}, nil
}

// SendEmail envía el correo usando las credenciales obtenidas de Secrets Manager.
func (s *SMTPEmailService) SendEmail(remitente, destinatarios, asunto, cuerpo string) error {
	// Validar la configuración SMTP
	if s.server == "" || s.port == "" || s.username == "" || s.password == "" {
		logs.LogError("Configuración SMTP incompleta en las variables de entorno", nil)
		return fmt.Errorf("error: configuración SMTP incompleta en las variables de entorno")
	}

	// Configurar autenticación SMTP
	auth := smtp.PlainAuth("", s.username, s.password, s.server)

	// Separar los destinatarios por comas
	to := strings.Split(destinatarios, ",")

	// Configurar mensaje en formato HTML
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		remitente,
		strings.Join(to, ", "),
		asunto,
		cuerpo,
	))

	// Manejar error de conversión de destinatarios
	if to == nil || len(to) == 0 || strings.Contains(to[0], "\x7f") {
		return fmt.Errorf("error al convertir destinatarios a JSON")
	}

	// Usar la función de envoltorio smtpSendMailWrapper
	err := s.sendMail(s.server+":"+s.port, auth, remitente, to, msg)
	if err != nil {
		return fmt.Errorf("error enviando el correo electrónico: %v", err)
	}

	return nil
}
