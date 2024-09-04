package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type AppConfig struct {
	Environment        string
	AWSRegion          string
	SQSQueueURL        string
	LocalstackEndpoint string
}

func LoadConfig() (*AppConfig, aws.Config) {
	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	environment := os.Getenv("ENVIRONMENT")
	awsRegion := os.Getenv("AWS_REGION")
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	localstackEndpoint := os.Getenv("LOCALSTACK_ENDPOINT")

	// Cargar la configuración de AWS con contexto y opciones
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	return &AppConfig{
		Environment:        environment,
		AWSRegion:          awsRegion,
		SQSQueueURL:        sqsQueueURL,
		LocalstackEndpoint: localstackEndpoint,
	}, cfg
}
