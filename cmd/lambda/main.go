package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/viper"
	"gmf_message_processor/config"
	"gmf_message_processor/connection"
	internalAws "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/handler"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/utils"
	"os"
	"time"
)

func main() {
	// Inicializar el Servicio de Secretos
	secretService := connection.NewSecretService()

	// Inicializar el Logger
	logger := &logs.LoggerAdapter{}

	// Inicializar la aplicación
	plantillaService, dbManager, err := config.InitApplication(secretService, "")
	if err != nil {
		logs.LogError("Error inicializando la aplicación", err, "")
		config.CleanupApplication(dbManager, "")
		return
	}

	// Función para cargar la configuración de AWS
	loadConfigFunc := func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (aws.Config, error) {
		return awsConfig.LoadDefaultConfig(ctx, optFns...)
	}

	// Inicializar SQSClient
	sqsClient, err := internalAws.NewSQSClient(viper.GetString("SQS_QUEUE_URL"), loadConfigFunc)
	if err != nil {
		logs.LogError("Error inicializando SQS Client", err, "")
		config.CleanupApplication(dbManager, "")
		return
	}

	// Crear una instancia de Utils (que implementa UtilsInterface)
	utilsImpl := &utils.Utils{}

	// Crear un nuevo manejador de SQS
	sqsHandler := handler.NewSQSHandler(plantillaService, sqsClient, utilsImpl, logger)

	// Verificar el entorno
	appEnv := viper.GetString("SERVICE_ENV")
	if appEnv == "local" {
		eventFilePath := "test_data/event.json"
		eventFile, err := os.ReadFile(eventFilePath)
		if err != nil {
			logs.LogError("Error al leer el archivo de evento", err, "")
			config.CleanupApplication(dbManager, "")
			return
		}

		// Deserializar el contenido del archivo event.json al tipo SQSEvent
		var sqsEvent events.SQSEvent
		err = json.Unmarshal(eventFile, &sqsEvent)
		if err != nil {
			logs.LogError("Error deserializando el archivo event.json", err, "")
			config.CleanupApplication(dbManager, "")
			return
		}

		// Extraer el message_id del evento
		messageID := sqsEvent.Records[0].MessageId

		// Log inicial de la aplicación con el message_id real
		logs.LogInfo("Inicia proceso de envío de correo", messageID)

		// Registrar el tiempo de inicio
		startTime := time.Now()

		// Procesar el evento simulado
		err = sqsHandler.HandleLambdaEvent(context.TODO(), sqsEvent)
		if err != nil {
			logger.LogError("Error procesando el evento SQS simulado", err, messageID)
		} else {
			duration := time.Since(startTime).Milliseconds() // Calcular la duración en milisegundos
			logger.LogInfo(fmt.Sprintf(
				"Fin ejecución proceso de envío de correo. status: EXITOSO, duración: %d ms",
				duration),
				messageID,
			)
		}
	} else {
		// Implementación para producción
		lambda.Start(func(ctx context.Context, sqsEvent events.SQSEvent) error {
			return sqsHandler.HandleLambdaEvent(ctx, sqsEvent)
		})
	}

	// Limpieza de recursos al finalizar
	config.CleanupApplication(dbManager, "")
}
