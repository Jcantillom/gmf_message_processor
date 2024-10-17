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
	"time"
)

var CleanupApplicationFunc = config.CleanupApplication

// FileReaderFunc define el tipo de una función que lee un archivo.
type FileReaderFunc func(filename string) ([]byte, error)

// ReadSQSEventFromFile lee y deserializa el archivo JSON de evento de SQS.
func ReadSQSEventFromFile(fileReader FileReaderFunc) (*events.SQSEvent, error) {
	// Leer el archivo JSON
	filePath := "test_data/event.json"
	eventFile, err := fileReader(filePath)
	if err != nil {
		logs.LogError("Error al leer el archivo de evento", err, "")
		return nil, err
	}

	// Deserializer el contenido del archivo event.json al tipo SQSEvent
	var sqsEvent events.SQSEvent
	err = json.Unmarshal(eventFile, &sqsEvent)
	if err != nil {
		logs.LogError("Error deserializando el archivo event.json", err, "")
		return nil, err
	}

	// Imprimir el evento SQS en formato JSON con sangría
	PrintEvent(sqsEvent)

	return &sqsEvent, nil
}

// PrintEvent imprime el evento SQS en formato JSON con sangría.
func PrintEvent(event events.SQSEvent) {
	eventJSON, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		fmt.Printf("Error al convertir el evento a JSON: %v\n", err)
		return
	}
	fmt.Printf("Evento SQS recibido:\n%s\n", string(eventJSON))
}

func ProcessLocalEvent(
	sqsHandler handler.SQSHandlerInterface,
	dbManager connection.DBManagerInterface,
	fileReader FileReaderFunc) {

	defer dbManager.CloseDB("")

	// Leer el archivo de evento SQS usando la función fileReader
	sqsEvent, err := ReadSQSEventFromFile(fileReader)
	if err != nil {
		CleanupApplicationFunc(dbManager, "")
		return
	}

	// Extraer el message_id del evento
	messageID := sqsEvent.Records[0].MessageId

	// Log inicial de la aplicación con el message_id real
	logs.LogInfo("Inicia proceso de envío de correo", messageID)

	// Registrar el tiempo de inicio
	startTime := time.Now()

	// Procesar el evento simulado
	err = sqsHandler.HandleLambdaEvent(context.TODO(), *sqsEvent)
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
