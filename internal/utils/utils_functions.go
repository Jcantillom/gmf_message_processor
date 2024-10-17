package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"os"
	"strconv"
	"strings"
)

type UtilsInterface interface {
	ExtractMessageBody(body, messageID string) (string, error)
	ValidateSQSMessage(messageBody string) (*models.SQSMessage, error)
	DeleteMessageFromQueue(
		ctx context.Context,
		client aws.SQSAPI,
		queueURL string,
		receiptHandle *string,
		messageID string) error
	SendMessageToQueue(
		ctx context.Context,
		client aws.SQSAPI,
		queueURL string,
		messageBody string,
		messageID string) error
}

type Utils struct{}

func (u *Utils) ExtractMessageBody(sqsBody string, messageID string) (string, error) {
	var message map[string]interface{}
	if err := json.Unmarshal([]byte(sqsBody), &message); err != nil {
		logs.LogError("Error deserializando el mensaje de SQS", err, messageID)
		return "", errors.New("error deserializando el mensaje de SQS")
	}
	return sqsBody, nil
}

func (u *Utils) DeleteMessageFromQueue(
	ctx context.Context, client aws.SQSAPI, queueURL string, receiptHandle *string, messageID string) error {
	_, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &queueURL,
		ReceiptHandle: receiptHandle,
	})

	if err != nil {
		logs.LogError("Error al eliminar el mensaje de SQS", err, messageID)
		return err
	}

	logs.LogDebug("Mensaje eliminado de SQS", messageID)

	return nil
}

// ValidateSQSMessage ...
func (u *Utils) ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	var msg models.SQSMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, errors.New("invalid JSON format")
	}
	if msg.IDPlantilla == "" {
		return nil, errors.New("id_plantilla is required")
	}
	return &msg, nil
}

// SendMessageToQueue ...
func (u *Utils) SendMessageToQueue(
	ctx context.Context, client aws.SQSAPI, queueURL string, messageBody string, messageID string) error {

	delaySecondsStr := os.Getenv("SQS_MESSAGE_DELAY")
	delaySeconds := 0 // Valor por defecto

	if delaySecondsStr != "" {
		var err error
		delaySeconds, err = strconv.Atoi(delaySecondsStr)
		if err != nil {
			logs.LogError("Error al convertir el valor de SQS_MESSAGE_DELAY", err, messageID)
			return err
		}
	}

	_, err := client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:     &queueURL,
		MessageBody:  &messageBody,
		DelaySeconds: int32(delaySeconds),
	})

	if err != nil {
		logs.LogError("Error al enviar el mensaje a SQS", err, messageID)
		return err
	}

	logs.LogInfo("Mensaje enviado a SQS con éxito", messageID)
	return nil
}

// ReplacePlaceholders ...
func ReplacePlaceholders(text string, params map[string]string) string {
	for key, value := range params {
		text = strings.ReplaceAll(text, key, value)
	}
	return text
}

// GetMaxRetries ...
func GetMaxRetries() int {
	maxRetriesStr := os.Getenv("MAX_RETRIES")
	if maxRetriesStr == "" {
		return 3
	}
	maxRetries, err := strconv.Atoi(maxRetriesStr)
	if err != nil {
		return 3
	}
	return maxRetries
}
