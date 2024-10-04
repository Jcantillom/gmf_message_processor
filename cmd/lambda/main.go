package main

import (
	"context"
	"github.com/spf13/viper"
	"gmf_message_processor/config"
	internalAws "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/handler"
	"gmf_message_processor/internal/logs"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	// Inicializar la aplicación y obtener el servicio necesario y el DBManager
	plantillaService, dbManager := config.InitApplication()

	// Función de carga de configuración predeterminada
	loadConfigFunc := func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (aws.Config, error) {
		return awsConfig.LoadDefaultConfig(ctx, optFns...)
	}

	// Inicializar SQSClient
	sqsClient, err := internalAws.NewSQSClient(viper.GetString("SQS_QUEUE_URL"), loadConfigFunc)
	if err != nil {
		logs.LogError(context.Background(), "Error inicializando SQS Client: %v", err)
		return
	}

	// Crear un nuevo manejador de SQS
	sqsHandler := handler.NewSQSHandler(plantillaService, sqsClient)

	// Procesar mensajes de SQS
	if err := sqsHandler.ProcessMessage(context.TODO()); err != nil {
		logs.LogError(context.Background(), "Error procesando mensajes de SQS ❌: %v", err)
	}

	// Limpieza de recursos al finalizar
	config.CleanupApplication(dbManager)
}
