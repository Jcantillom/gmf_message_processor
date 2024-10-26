package email

import (
	"fmt"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/logs"
	"gopkg.in/gomail.v2"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// DialerInterface permite la inyección de un mock del dialer en los tests.
type DialerInterface interface {
	DialAndSend(m ...*gomail.Message) error
}
type LoggerInterface interface {
	LogError(message string, err error, messageID string)
	LogInfo(message string, messageID string)
}

// SMTPEmailService implementa EmailService utilizando SMTP.
type SMTPEmailService struct {
	server   string
	port     string
	username string
	password string
	dialer   DialerInterface
	timeout  time.Duration
	logger   LoggerInterface
}

// NewSMTPEmailService crea una instancia de SMTPEmailService usando SecretService.
func NewSMTPEmailService(
	secretService connection.SecretService, messageID string, logger LoggerInterface,
) (*SMTPEmailService, error) {
	secretName := os.Getenv("SECRETS_SMTP")
	secretData, err := secretService.GetSecret(secretName, messageID)
	if err != nil {
		logger.LogError("Error al obtener las credenciales SMTP desde Secrets Manager", err, messageID)
		return nil, err
	}

	timeoutStr := viper.GetString("SMTP_TIMEOUT")
	timeoutValue, err := strconv.Atoi(timeoutStr)
	if err != nil {
		timeoutValue = 15
	}

	server := os.Getenv("SMTP_SERVER")
	port := os.Getenv("SMTP_PORT")

	dialer := gomail.NewDialer(server, viper.GetInt("SMTP_PORT"), secretData.Username, secretData.Password)

	return &SMTPEmailService{
		server:   server,
		port:     port,
		username: secretData.Username,
		password: secretData.Password,
		dialer:   dialer,
		timeout:  time.Duration(timeoutValue) * time.Second,
		logger:   logger, // Usamos el logger inyectable aquí
	}, nil
}

var stat = os.Stat

// SendEmail envía un correo usando el dialer configurado.
func (s *SMTPEmailService) SendEmail(
	remitente, destinatarios, asunto, cuerpo, imagePath, imageName, messageID string,
) error {
	if s.server == "" || s.port == "" || s.username == "" || s.password == "" {
		return fmt.Errorf("error: configuración SMTP incompleta en las variables de entorno")
	}

	// Verificar que los destinatarios no estén vacíos
	if strings.TrimSpace(destinatarios) == "" {
		return fmt.Errorf("error: no se especificaron destinatarios")
	}

	// Intentar dividir los destinatarios y verificar si son válidos
	toAddresses := strings.Split(destinatarios, ",")
	for _, addr := range toAddresses {
		if strings.TrimSpace(addr) == "" || !strings.Contains(addr, "@") {
			return fmt.Errorf("error al convertir destinatarios a JSON")
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", remitente)
	m.SetHeader("To", toAddresses...)
	m.SetHeader("Subject", asunto)
	m.SetBody("text/html", cuerpo)

	if _, err := stat(imagePath); os.IsNotExist(err) {
		s.logger.LogError(fmt.Sprintf(
			"La imagen %s no existe en la ruta proporcionada", imagePath), err, messageID)
		return err
	}
	m.Embed(imagePath, gomail.SetHeader(map[string][]string{
		"Content-ID": {"<" + imageName + ">"},
	}))

	if err := s.dialer.DialAndSend(m); err != nil {
		logs.LogError("Error al enviar el correo electrónico", err, messageID)
		return err
	}

	logs.LogInfo("Correo enviado con éxito", messageID)
	return nil
}
