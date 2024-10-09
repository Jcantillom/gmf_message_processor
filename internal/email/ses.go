package email

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"gmf_message_processor/internal/logs"
	"net/mail"
	"strings"
)

// SESClientInterface define los métodos que se utilizarán del cliente SES.
type SESClientInterface interface {
	SendEmail(
		ctx context.Context,
		input *ses.SendEmailInput, opts ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

// SESEmailService implementa EmailService utilizando Amazon SES.
type SESEmailService struct {
	Client SESClientInterface // Usamos la interfaz SESClientInterface
	Sender string
}

// AWSConfigLoader es una interfaz que representa la función para cargar la configuración de AWS.
type AWSConfigLoader interface {
	LoadConfig(ctx context.Context) (aws.Config, error)
}

// RealAWSConfigLoader implementa AWSConfigLoader y usa la función real de aws-sdk-go-v2.
type RealAWSConfigLoader struct{}

// LoadConfig llama a la función real awsConfig.LoadDefaultConfig.
func (l *RealAWSConfigLoader) LoadConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

// NewSESEmailService crea una nueva instancia de SESEmailService.
func NewSESEmailService(configLoader AWSConfigLoader) *SESEmailService {
	cfg, err := configLoader.LoadConfig(context.TODO())
	if err != nil {
		logs.LogError("Error cargando la configuración de AWS: %v", err)
		return nil
	}

	client := ses.NewFromConfig(cfg)

	return &SESEmailService{
		Client: client,
	}
}

// Validar correos electrónicos
func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	return err
}

// SendEmail envía un correo electrónico usando SES.
func (s *SESEmailService) SendEmail(remitente, destinatarios, asunto, cuerpo string) error {
	// Validar correo del remitente
	remitente = strings.TrimSpace(remitente)
	if err := ValidateEmail(remitente); err != nil {
		logs.LogError("Correo remitente inválido: %v", err)
		return fmt.Errorf("correo remitente inválido: %v", err)
	}

	// Validar correos de los destinatarios
	destinatarios = strings.TrimSpace(destinatarios)
	for _, destinatario := range strings.Split(destinatarios, ",") {
		destinatario = strings.TrimSpace(destinatario)
		if err := ValidateEmail(destinatario); err != nil {
			logs.LogError("Correo destinatario inválido: %v", err)
			return fmt.Errorf("correo destinatario inválido: %v", err)
		}
	}

	// Crear la solicitud para enviar el correo
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: strings.Split(destinatarios, ","),
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(cuerpo),
				},
			},
			Subject: &types.Content{
				Data: aws.String(asunto),
			},
		},
		Source: aws.String(remitente),
	}

	// Enviar el correo utilizando SES
	_, err := s.Client.SendEmail(context.TODO(), input)
	if err != nil {
		logs.LogError("Error enviando el correo electrónico con SES: %v", err)
		return fmt.Errorf("error enviando correo con SES: %v", err)
	}

	return nil
}
