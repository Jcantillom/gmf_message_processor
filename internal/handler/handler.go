package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
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
	SendMessageToQueue(
		ctx context.Context,
		client aws.SQSAPI,
		queueURL string,
		messageBody string,
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

		// Eliminar el mensaje de SQS antes de procesarlo
		err = h.Utils.DeleteMessageFromQueue(
			ctx, h.SQSClient, h.SQSClient.GetQueueURL(), &record.ReceiptHandle, messageID)
		if err != nil {
			h.Logger.LogError("Error al eliminar el mensaje de SQS", err, messageID)
			return err
		}

		// Manejar el procesamiento dentro de un bloque de recuperación
		defer func() {
			if r := recover(); r != nil {
				h.Logger.LogError("Error al procesar el mensaje", fmt.Errorf("%v", r), messageID)
				validMsg.RetryCount++
				if validMsg.RetryCount <= utils.GetMaxRetries() { // asegúrate de que no exceda el máximo
					h.Logger.LogInfo(fmt.Sprintf(
						"Reintentando el mensaje. Conteo actual: %d", validMsg.RetryCount), messageID)

					// Preparar el nuevo cuerpo del mensaje con el contador de reintentos
					messageBodyWithRetry := fmt.Sprintf(
						"{\"id_plantilla\": \"%s\", \"parametros\": %s, \"retry_count\": %d}",
						validMsg.IDPlantilla,
						validMsg.Parametro,
						validMsg.RetryCount,
					)
					// Vuelve a enviar el mensaje a la cola
					if err := h.Utils.SendMessageToQueue(
						ctx, h.SQSClient, h.SQSClient.GetQueueURL(), messageBodyWithRetry, messageID); err != nil {
						h.Logger.LogError("Error al reenviar el mensaje a SQS", err, messageID)
					}
				} else {
					h.Logger.LogError("Se alcanzó el máximo de reintentos", nil, messageID)
				}
			}
		}()

		if err := h.PlantillaService.HandlePlantilla(ctx, validMsg, messageID); err != nil {
			h.Logger.LogError("Error al procesar el mensaje", err, messageID)

			// Aquí puedes enviar el mensaje nuevamente a SQS
			validMsg.RetryCount++ // aumentar el contador de reintentos
			if validMsg.RetryCount <= utils.GetMaxRetries() {
				h.Logger.LogInfo(fmt.Sprintf(
					"Reintentando el mensaje. Conteo actual: %d", validMsg.RetryCount), messageID)

				// Preparar el nuevo cuerpo del mensaje con el contador de reintentos
				messageBodyWithRetry := fmt.Sprintf(
					"{\"id_plantilla\": \"%s\", \"parametros\": %s, \"retry_count\": %d}",
					validMsg.IDPlantilla,
					// Aquí debes agregar la lógica para convertir validMsg.Parametro a un JSON válido
					validMsg.Parametro,
					validMsg.RetryCount,
				)

				if err := h.Utils.SendMessageToQueue(
					ctx, h.SQSClient, h.SQSClient.GetQueueURL(), messageBodyWithRetry, messageID); err != nil {
					h.Logger.LogError(
						"Error al reenviar el mensaje a SQS", err, messageID)
				}
			} else {
				h.Logger.LogError(
					"Se alcanzó el máximo de reintentos", nil, messageID)
			}
		}
	}
	return nil
}
