package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/viper"
	"gmf_message_processor/config"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/local"
	"os"
)

func main() {
	// Inicializar servicios
	secretService := connection.NewSecretService()
	// Inicializar el DBManager
	dbManager := connection.NewDBManager(secretService)

	// Inicializar la aplicación
	appContext, err := config.InitApplication(
		"",
		secretService,
		dbManager,
	)
	if err != nil {
		logs.LogError("Error inicializando la aplicación", err, "")
		config.CleanupApplication(appContext.DBManager, "")
		return
	}

	// Verificar el entorno de ejecución para determinar si se ejecuta localmente o en AWS.
	appEnv := viper.GetString("SERVICE_ENV")
	if appEnv == "local" {
		local.ProcessLocalEvent(
			appContext.SQSHandler,
			appContext.DBManager,
			os.ReadFile,
		)

	} else {
		// Implementación para producción
		lambda.Start(func(ctx context.Context, sqsEvent events.SQSEvent) error {
			return appContext.SQSHandler.HandleLambdaEvent(ctx, sqsEvent)
		})
	}

	// Limpieza de recursos al finalizar
	config.CleanupApplication(appContext.DBManager, "")
}
