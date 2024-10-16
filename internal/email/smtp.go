package email

import (
	"context"
	"fmt"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/logs"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// EmailServiceInterface define los métodos que debe implementar un servicio de correo electrónico.
type EmailServiceInterface interface {
	SendEmail(remitente, destinatarios, asunto, cuerpo, messageID string) error
}

var ctx = context.TODO()

// SMTPEmailService implementa EmailService utilizando SMTP.
type SMTPEmailService struct {
	server   string
	port     string
	username string
	password string
	sendMail smtpSendMailFunc
	timeout  time.Duration
}

// smtpSendMailFunc es una función de envío de correo electrónico SMTP.
type smtpSendMailFunc func(
	addr string, a smtp.Auth, from string, to []string, msg []byte) error

// NewSMTPEmailService crea una nueva instancia de SMTPEmailService usando SecretService para obtener las credenciales SMTP.
func NewSMTPEmailService(secretService connection.SecretService, messageID string) (*SMTPEmailService, error) {
	secretName := os.Getenv("SECRETS_SMTP")
	secretData, err := secretService.GetSecret(secretName, messageID) // Pasar el messageID
	if err != nil {
		logs.LogError("Error al obtener las credenciales SMTP desde Secrets Manager", err, messageID)
		return nil, err
	}

	// Leer el timeout desde las variables de entorno, o usar el valor por defecto (15 segundos)
	timeoutStr := viper.GetString("SMTP_TIMEOUT")
	timeoutValue, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeoutValue = 15 // Valor por defecto en segundos
	}

	return &SMTPEmailService{
		server:   os.Getenv("SMTP_SERVER"),
		port:     os.Getenv("SMTP_PORT"),
		username: secretData.Username,
		password: secretData.Password,
		sendMail: smtp.SendMail,
		timeout:  time.Duration(timeoutValue) * time.Second, // Convertir a duración en segundos
	}, nil
}

// SendEmail envía el correo con el timeout configurable.
func (s *SMTPEmailService) SendEmail(
	remitente,
	destinatarios,
	asunto,
	cuerpo string,
	messageID string) error {
	// Validar la configuración SMTP
	if s.server == "" || s.port == "" || s.username == "" || s.password == "" {
		return fmt.Errorf("error: configuración SMTP incompleta en las variables de entorno")
	}

	// Configurar autenticación SMTP
	auth := smtp.PlainAuth("", s.username, s.password, s.server)

	// Separar los destinatarios por comas
	to := strings.Split(destinatarios, ",")

	// validar que haya destinatarios
	if len(to) == 0 || to[0] == "" {
		return fmt.Errorf("error: no se especificaron destinatarios")
	}

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

	// Registrar el inicio del envío de correo
	logs.LogInfo("Inicia consumo de HOST_SMTP externo para envío de correo", messageID)

	// Medir el tiempo de inicio
	startTime := time.Now()

	// Enviar el correo con el timeout configurado
	err := s.sendMailWithTimeout(s.server+":"+s.port, auth, remitente, to, msg)

	// Medir el tiempo de fin
	duration := time.Since(startTime).Milliseconds()

	// Registrar la duración del envío de correo
	if err != nil {
		logs.LogError(fmt.Sprintf(
			"Fin consumo de HOST_SMTP externo para envío de correo, duración %d ms, error: %v",
			duration, err),
			err, messageID,
		)
		return err
	}

	logs.LogInfo(fmt.Sprintf(
		"Fin consumo de HOST_SMTP externo para envío de correo, duración %d ms",
		duration),
		messageID,
	)

	return nil
}

// sendMailWithTimeout envía el correo con un timeout establecido.
func (s *SMTPEmailService) sendMailWithTimeout(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		err := s.sendMail(addr, auth, from, to, msg)
		done <- err
	}()

	select {
	case <-ctx.Done():
		// Timeout alcanzado
		return fmt.Errorf("error: timeout al enviar correo electrónico")
	case err := <-done:
		// Error al enviar el correo
		if err != nil {
			return fmt.Errorf("error enviando el correo electrónico: %v", err)
		}
		return nil
	}
}
