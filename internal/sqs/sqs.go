package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/jcantillom/gmf_message_processor/config"
	"github.com/jcantillom/gmf_message_processor/internal/processor"
	"log"
)

type SQSClient struct {
	client *sqs.Client
	config *config.AppConfig
}

func NewSQSClient(cfg *config.AppConfig, awsConfig aws.Config) *SQSClient {
	// Configurar el endpoint en función del entorno
	if cfg.Environment == "local" {
		// Usar LocalStack para entorno local
		localstackEndpoint := cfg.LocalstackEndpoint
		awsConfig.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: localstackEndpoint, // URL de LocalStack desde el archivo .env
				}, nil
			})
	}
	// En entorno de producción, usar el endpoint por defecto de AWS

	return &SQSClient{
		client: sqs.NewFromConfig(awsConfig),
		config: cfg,
	}
}

func (s *SQSClient) ReceiveMessages() {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &s.config.SQSQueueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     10,
	}

	result, err := s.client.ReceiveMessage(context.TODO(), input)
	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}

	for _, message := range result.Messages {
		// Validar el mensaje JSON para asegurar que tenga el formato correcto
		validMessage, err := processor.ValidateMessage(*message.Body)
		if err != nil {
			log.Printf("Error: %v\n", err)
			continue // Saltar al siguiente mensaje si el formato no es válido
		}

		// Formatear el mensaje JSON para que sea más legible
		formattedJSON, err := json.MarshalIndent(validMessage, "", "  ")
		if err != nil {
			log.Fatalf("Error formatting JSON: %v", err)
		}

		fmt.Printf("Message received 🚀:\n%s\n", string(formattedJSON))
		// Aquí puedes llamar a una función para procesar el mensaje.
	}
}
