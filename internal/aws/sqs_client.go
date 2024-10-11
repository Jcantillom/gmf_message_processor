package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
)

// SQSAPI define una interfaz para las operaciones de SQS.
type SQSAPI interface {
	DeleteMessage(
		ctx context.Context,
		input *sqs.DeleteMessageInput,
		opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
	GetQueueURL() string
	SendMessage(
		ctx context.Context,
		input *sqs.SendMessageInput,
		opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// SQSClient define la estructura del cliente de SQS.
type SQSClient struct {
	Client   *sqs.Client
	QueueURL string
}

// GetQueueURL devuelve la URL de la cola SQS.
func (s *SQSClient) GetQueueURL() string {
	return s.QueueURL
}

// DeleteMessage elimina un mensaje de la cola SQS.
func (s *SQSClient) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	return s.Client.DeleteMessage(ctx, input, opts...)
}

// SendMessage envía un mensaje a la cola SQS.
func (s *SQSClient) SendMessage(ctx context.Context, input *sqs.SendMessageInput, opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	return s.Client.SendMessage(ctx, input, opts...)
}

// NewSQSClient inicializa un nuevo cliente de SQS.
func NewSQSClient(
	queueURL string, loadConfigFunc func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)) (*SQSClient, error) {
	// Validar la URL de la cola
	if err := validateQueueURL(queueURL); err != nil {
		return nil, err
	}

	// Obtener la configuración de AWS
	cfg, err := loadConfigFunc(context.TODO(), config.WithEndpointResolver(getEndpointResolver()))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Crear el cliente SQS
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.HTTPClient = &http.Client{}
	})

	return &SQSClient{
		Client:   client,
		QueueURL: queueURL,
	}, nil
}

// validateQueueURL valida la URL de la cola SQS.
func validateQueueURL(queueURL string) error {
	if _, err := url.ParseRequestURI(queueURL); err != nil {
		return fmt.Errorf("invalid queue URL: %v", err)
	}
	return nil
}

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
				return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested for service %s in region %s", service, region)
			default:
				return aws.Endpoint{}, fmt.Errorf("unknown APP_ENV: %s", appEnv) // Asegúrate de que esto esté incluido
			}
		}
		return aws.Endpoint{}, fmt.Errorf("unknown service: %s", service)
	})
}
