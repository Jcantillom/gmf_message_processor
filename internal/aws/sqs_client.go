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
	ReceiveMessage(
		ctx context.Context, input *sqs.ReceiveMessageInput,
		opts ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(
		ctx context.Context, input *sqs.DeleteMessageInput,
		opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// SQSClient define la estructura del cliente de SQS.
type SQSClient struct {
	Client   SQSAPI
	QueueURL string
}

// NewSQSClient inicializa un nuevo cliente de SQS.
func NewSQSClient(
	queueURL string, loadConfigFunc func(ctx context.Context,
		optFns ...func(*config.LoadOptions) error) (aws.Config, error)) (*SQSClient, error) {
	// Validar la URL de la cola
	if err := validateQueueURL(queueURL); err != nil {
		return nil, err
	}

	// Obtener el endpoint resolver basado en el entorno
	endpointResolver, err := getEndpointResolver()
	if err != nil {
		return nil, err
	}

	// Cargar la configuración de AWS
	cfg, err := loadConfigFunc(context.TODO(), config.WithEndpointResolver(endpointResolver))
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

// getEndpointResolver obtiene el resolvedor de endpoints basado en el entorno de la aplicación.
func getEndpointResolver() (aws.EndpointResolver, error) {
	appEnv := viper.GetString("APP_ENV")

	switch appEnv {
	case "local":
		return aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == sqs.ServiceID && region == "us-east-1" {
				return aws.Endpoint{
					URL:           "http://localhost:4566",
					SigningRegion: "us-east-1",
				}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		}), nil
	case "dev", "qa", "prod":
		return aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		}), nil
	default:
		return nil, fmt.Errorf("unknown APP_ENV: %s", appEnv)
	}
}
