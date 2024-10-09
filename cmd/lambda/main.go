package main

import (
	"context"
	"github.com/spf13/viper"
	"gmf_message_processor/config"
	"gmf_message_processor/connection"
	internalAws "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/handler"
	"gmf_message_processor/internal/logs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	// Inicializar el Servicio de Secretos
	secretService := connection.NewSecretService()

	// Inicializar la aplicación
	plantillaService, dbManager, _ := config.InitApplication(secretService)

	// Función de carga de configuración predeterminada
	loadConfigFunc := func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (aws.Config, error) {
		return awsConfig.LoadDefaultConfig(ctx, optFns...)
	}

	// Inicializar SQSClient
	sqsClient, err := internalAws.NewSQSClient(viper.GetString("SQS_QUEUE_URL"), loadConfigFunc)
	if err != nil {
		logs.LogError("Error inicializando SQS Client: %v", err)
		// Al producirse un error crítico al inicializar el cliente, llamamos a CleanupApplication antes de salir
		config.CleanupApplication(dbManager)
		return
	}

	// Crear un nuevo manejador de SQS
	sqsHandler := handler.NewSQSHandler(plantillaService, sqsClient)

	// Procesar mensajes de SQS
	if err := sqsHandler.ProcessMessage(context.TODO()); err != nil {
		// Llamamos a CleanupApplication para asegurarnos de que los recursos se liberen correctamente
		config.CleanupApplication(dbManager)
		return
	}

	// Limpieza de recursos al finalizar
	config.CleanupApplication(dbManager)
}
