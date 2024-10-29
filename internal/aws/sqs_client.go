package aws

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	messageInvalidURL = "invalid queue URL: %v"
)

// SQSAPI define una interfaz para las operaciones de SQS.
type SQSAPI interface {
	DeleteMessage(
		ctx context.Context,
		input *sqs.DeleteMessageInput,
		opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	SendMessage(
		ctx context.Context,
		input *sqs.SendMessageInput,
		opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// SQSClient define la estructura del cliente de SQS.
type SQSClient struct {
	Client   SQSAPI
	QueueURL string
}

// GetQueueURL devuelve la URL de la cola SQS.
func (s *SQSClient) GetQueueURL() string {
	return s.QueueURL
}

// newSecureHTTPClient crea un cliente HTTP seguro con configuraciones específicas.
func newSecureHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second, // Limitar el tiempo máximo de las solicitudes
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // Asegurar el uso de TLS 1.2 o superior
			},
			DisableKeepAlives:     false,
			MaxIdleConns:          10, // Limitar las conexiones ociosas
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
}

// DeleteMessage elimina un mensaje de la cola SQS con validaciones adicionales.
func (s *SQSClient) DeleteMessage(
	ctx context.Context,
	input *sqs.DeleteMessageInput,
	opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {

	// Validar y sanitizar el input antes de ejecutar la solicitud.
	if err := sanitizeAndValidateInput(input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Ejecutar la operación de DeleteMessage de forma segura.
	output, err := s.Client.DeleteMessage(ctx, input, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to delete message: %w", err)
	}

	return output, nil
}

// sanitizeAndValidateInput aplica validaciones al input.
func sanitizeAndValidateInput(input *sqs.DeleteMessageInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.ReceiptHandle == nil || *input.ReceiptHandle == "" {
		return fmt.Errorf("receipt handle is required and cannot be empty")
	}
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return fmt.Errorf("queue URL is required and cannot be empty")
	}

	// Validar la URL.
	if err := validateQueueURL(*input.QueueUrl); err != nil {
		return fmt.Errorf(messageInvalidURL, err)
	}

	// No modificar el ReceiptHandle.
	return nil
}

// sanitizeString aplica una sanitización más estricta.
func sanitizeString(input string) string {
	// Escapar caracteres no permitidos en las entradas.
	return url.QueryEscape(strings.TrimSpace(input))
}

// validateQueueURL valida la URL de la cola SQS.
func validateQueueURL(queueURL string) error {
	parsedURL, err := url.ParseRequestURI(queueURL)
	if err != nil {
		return fmt.Errorf(messageInvalidURL, err)
	}

	appEnv := viper.GetString("APP_ENV")

	// Permitir HTTP solo en entornos locales
	if parsedURL.Scheme != "https" && appEnv != "local" {
		return fmt.Errorf("only HTTPS URLs are allowed for non-local environments")
	}

	return nil
}

// SendMessage envía un mensaje a la cola SQS.
func (s *SQSClient) SendMessage(
	ctx context.Context,
	input *sqs.SendMessageInput,
	opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {

	// Validar el input para SendMessage.
	if err := validateSendMessageInput(input); err != nil {
		return nil, err
	}

	return s.Client.SendMessage(ctx, input, opts...)
}

// validateSendMessageInput valida el input del mensaje.
func validateSendMessageInput(input *sqs.SendMessageInput) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}
	if input.QueueUrl == nil || *input.QueueUrl == "" {
		return fmt.Errorf("queue URL is required and cannot be empty")
	}
	if input.MessageBody == nil || *input.MessageBody == "" {
		return fmt.Errorf("message body is required and cannot be empty")
	}
	if err := validateQueueURL(*input.QueueUrl); err != nil {
		return fmt.Errorf(messageInvalidURL, err)
	}
	return nil
}

// NewSQSClient inicializa un nuevo cliente de SQS con un cliente HTTP seguro.
func NewSQSClient(
	queueURL string,
	loadConfigFunc func(
		ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)) (*SQSClient, error) {

	// Validar la URL de la cola.
	if err := validateQueueURL(queueURL); err != nil {
		return nil, fmt.Errorf("invalid queue URL provided: %w", err)
	}

	// Cargar la configuración de AWS.
	cfg, err := loadConfigFunc(context.TODO(), config.WithEndpointResolver(getEndpointResolver()))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Crear el cliente SQS utilizando un cliente HTTP seguro.
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.HTTPClient = newSecureHTTPClient() // Aplicar cliente HTTP seguro.
	})

	return &SQSClient{
		Client:   client,
		QueueURL: queueURL,
	}, nil
}

// getEndpointResolver devuelve un resolutor de endpoint según el entorno.
func getEndpointResolver() aws.EndpointResolver {
	appEnv := viper.GetString("APP_ENV")

	return aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == sqs.ServiceID {
			switch appEnv {
			case "local":
				return aws.Endpoint{
					URL:           "http://localhost:4566",
					SigningRegion: "us-east-1",
				}, nil
			case "dev", "prod", "qa":
				return aws.Endpoint{}, fmt.Errorf(
					"unknown endpoint requested for service %s in region %s", service, region)
			default:
				return aws.Endpoint{}, fmt.Errorf("unknown APP_ENV: %s", appEnv)
			}
		}
		return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
	})
}
