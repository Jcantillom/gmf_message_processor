package utils

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"strings"
)

type UtilsInterface interface {
	ExtractMessageBody(body, messageID string) (string, error)
	ValidateSQSMessage(messageBody string) (*models.SQSMessage, error)
	DeleteMessageFromQueue(
		ctx context.Context,
		client aws.SQSAPI, // Cambiado a aws.SQSAPI
		queueURL string,
		receiptHandle *string,
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

	return nil
}

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

// ReplacePlaceholders ...
func ReplacePlaceholders(text string, params map[string]string) string {
	for key, value := range params {
		text = strings.ReplaceAll(text, key, value)
	}
	return text
}
