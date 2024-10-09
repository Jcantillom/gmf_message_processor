package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/logs"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

// ReadSQSEventFromFile lee un archivo JSON y lo parsea a un evento de SQS.
func ReadSQSEventFromFile(filePath string) (events.SQSEvent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return events.SQSEvent{}, err
	}

	var sqsEvent events.SQSEvent
	if err := json.Unmarshal(data, &sqsEvent); err != nil {
		return events.SQSEvent{}, err
	}

	return sqsEvent, nil
}

func ExtractMessageBody(sqsBody string) (string, error) {
	// Verificar si el mensaje es un JSON válido
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(sqsBody), &message); err != nil {
		return "", errors.New("error deserializando el mensaje de SQS")
	}

	// Retornar el mensaje deserializado como string, si es necesario
	return sqsBody, nil
}

// DeleteMessageFromQueue elimina un mensaje de la cola SQS.
func DeleteMessageFromQueue(ctx context.Context, client SQSAPI, queueURL string, receiptHandle *string) error {
	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		logs.LogError("Error al eliminar el mensaje de SQS: %v", err)
		return err
	}
	logs.LogInfo("Mensaje eliminado de la cola de SQS.")
	return nil
}
