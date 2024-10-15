package local

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"gmf_message_processor/config"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/handler"
	"gmf_message_processor/internal/logs"
	"os"
	"time"
)

func ProcessLocalEvent(sqsHandler *handler.SQSHandler, dbManager *connection.DBManager) {
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
		logs.LogError("Error procesando el evento SQS simulado", err, messageID)
	} else {
		duration := time.Since(startTime).Milliseconds() // Calcular la duración en milisegundos
		logs.LogInfo(
			fmt.Sprintf("Fin ejecución proceso de envío de correo. status: EXITOSO, duración: %d ms", duration),
			messageID,
		)
	}
}
