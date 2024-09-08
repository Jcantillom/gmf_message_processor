package email

import (
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// SendEmail envía un correo electrónico a los destinatarios proporcionados.
func SendEmail(remitente, destinatarios, asunto, cuerpo string) error {
	// Cargar configuración SMTP desde variables de entorno
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpServer == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		return fmt.Errorf("error: configuración SMTP incompleta en las variables de entorno")
	}

	// Configurar autenticación SMTP
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpServer)

	// Separar los destinatarios por comas
	to := strings.Split(destinatarios, ",")

	// Convertir la lista de destinatarios a JSON para un log estructurado
	toJSON, err := json.MarshalIndent(to, "", "  ")
	if err != nil {
		return fmt.Errorf("error al convertir destinatarios a JSON: %v", err)
	}

	log.Printf("Enviando correo a los siguientes destinatarios 📧:\n%s", toJSON)

	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", strings.Join(to, ", "), asunto, cuerpo))

	// Enviar el correo electrónico
	err = smtp.SendMail(smtpServer+":"+smtpPort, auth, remitente, to, msg)
	if err != nil {
		return fmt.Errorf("error enviando el correo electrónico: %v", err)
	}

	log.Printf("Correo electrónico enviado con éxito a 📤:\n%s", toJSON)

	return nil
}
