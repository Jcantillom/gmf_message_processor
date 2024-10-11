package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/models"
)

// LogInterface ...
type LogInterface interface {
	LogError(message string, err error, messageID string)
	LogInfo(message, messageID string)
}

// PlantillaServiceInterface ...
type PlantillaServiceInterface interface {
	HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error
}

// UtilsInterface ...
type UtilsInterface interface {
	ExtractMessageBody(body, messageID string) (string, error)
	ValidateSQSMessage(messageBody string) (*models.SQSMessage, error)
	DeleteMessageFromQueue(
		ctx context.Context,
		client aws.SQSAPI,
		queueURL string,
		receiptHandle *string,
		messageID string) error
}

// SQSHandler ...
type SQSHandler struct {
	PlantillaService PlantillaServiceInterface
	SQSClient        aws.SQSAPI
	Utils            UtilsInterface
	Logger           LogInterface
}

// NewSQSHandler ...
func NewSQSHandler(
	plantillaService PlantillaServiceInterface,
	sqsClient aws.SQSAPI,
	utils UtilsInterface,
	logger LogInterface) *SQSHandler {
	return &SQSHandler{
		PlantillaService: plantillaService,
		SQSClient:        sqsClient,
		Utils:            utils,
		Logger:           logger,
	}
}

// HandleLambdaEvent ...
func (h *SQSHandler) HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, record := range sqsEvent.Records {
		messageID := record.MessageId

		messageBody, err := h.Utils.ExtractMessageBody(record.Body, messageID)
		if err != nil {
			h.Logger.LogError("Error extrayendo el cuerpo del mensaje", err, messageID)
			return err
		}

		validMsg, err := h.Utils.ValidateSQSMessage(messageBody)
		if err != nil {
			h.Logger.LogError("Error validando el mensaje", err, messageID)
			return err
		}

		if err := h.PlantillaService.HandlePlantilla(ctx, validMsg, messageID); err != nil {
			h.Logger.LogError("Error al procesar el mensaje", err, messageID)
			return err
		}

		err = h.Utils.DeleteMessageFromQueue(
			ctx, h.SQSClient,
			h.SQSClient.GetQueueURL(),
			&record.ReceiptHandle,
			messageID)
		if err != nil {
			h.Logger.LogError("Error al eliminar el mensaje de SQS", err, messageID)
			return err
		}
	}
	return nil
}
