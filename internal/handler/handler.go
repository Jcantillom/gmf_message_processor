package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
)

// SQSHandlerInterface ...
type SQSHandlerInterface interface {
	HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error
}

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
	QueueURL         string // Nueva propiedad para almacenar la URL de la cola
}

// NewSQSHandler ...
func NewSQSHandler(
	plantillaService PlantillaServiceInterface,
	sqsClient aws.SQSAPI,
	utils UtilsInterface,
	logger LogInterface,
	queueURL string) *SQSHandler { // Recibe la QueueURL como parámetro
	return &SQSHandler{
		PlantillaService: plantillaService,
		SQSClient:        sqsClient,
		Utils:            utils,
		Logger:           logger,
		QueueURL:         queueURL, // Asigna la URL de la cola
	}
}

// HandleLambdaEvent ...
func (h *SQSHandler) HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error {

	// Imprimir el evento SQS completo con formato JSON
	printSQSEvent(sqsEvent)
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
			ctx, h.SQSClient, h.QueueURL, &record.ReceiptHandle, messageID) // Usa h.QueueURL en lugar de GetQueueURL()
		if err != nil {
			h.Logger.LogError("Error al eliminar el mensaje de SQS", err, messageID)
			return err
		}

		// Manejar el procesamiento dentro de un bloque de recuperación
		defer func() {
			if r := recover(); r != nil {
				h.Logger.LogError("Error al procesar el mensaje", fmt.Errorf("%v", r), messageID)
				validMsg.RetryCount++
				if validMsg.RetryCount <= utils.GetMaxRetries() {
					h.Logger.LogInfo(fmt.Sprintf(
						"Reintentando el mensaje. Conteo actual: %d", validMsg.RetryCount), messageID)

					// Preparar el nuevo cuerpo del mensaje con el contador de reintentos
					messageBodyWithRetry, err := json.Marshal(struct {
						IDPlantilla string                 `json:"id_plantilla"`
						Parametros  []models.ParametrosSQS `json:"parametros"`
						RetryCount  int                    `json:"retry_count"`
					}{
						IDPlantilla: validMsg.IDPlantilla,
						Parametros:  validMsg.Parametro,
						RetryCount:  validMsg.RetryCount,
					})

					if err != nil {
						h.Logger.LogError("Error al convertir el mensaje a JSON", err, messageID)
						return
					}

					// Imprimir el mensaje con el reintento para verificar su contenido
					fmt.Printf("Mensaje con reintento:\n%s\n", string(messageBodyWithRetry))

					// Enviar el mensaje nuevamente a SQS
					if err := h.Utils.SendMessageToQueue(
						ctx, h.SQSClient, h.QueueURL, string(messageBodyWithRetry), messageID); err != nil {
						h.Logger.LogError("Error al reenviar el mensaje a SQS", err, messageID)
					}
				} else {
					h.Logger.LogError("Se alcanzó el máximo de reintentos", nil, messageID)
				}
			}
		}()

		if err := h.PlantillaService.HandlePlantilla(ctx, validMsg, messageID); err != nil {
			h.Logger.LogError("Error al procesar el mensaje", err, messageID)

			// Aumentar el contador de reintentos
			validMsg.RetryCount++ // aumentar el contador de reintentos

			// Comprobar si no excede el máximo permitido
			if validMsg.RetryCount <= utils.GetMaxRetries() {
				h.Logger.LogInfo(fmt.Sprintf(
					"Reintentando el mensaje. Conteo actual: %d", validMsg.RetryCount), messageID)

				// Preparar el nuevo cuerpo del mensaje con el contador de reintentos
				messageBodyWithRetry, err := json.Marshal(struct {
					IDPlantilla string                 `json:"id_plantilla"`
					Parametros  []models.ParametrosSQS `json:"parametros"`
					RetryCount  int                    `json:"retry_count"`
				}{
					IDPlantilla: validMsg.IDPlantilla,
					Parametros:  validMsg.Parametro,
					RetryCount:  validMsg.RetryCount,
				})
				fmt.Printf("Contenido de Parametros: %+v\n", validMsg.Parametro)

				if err != nil {
					h.Logger.LogError("Error al convertir el mensaje a JSON", err, messageID)
					return err
				}

				// Imprimir el JSON del mensaje para verificar su contenido
				fmt.Printf("Mensaje con reintento:\n%s\n", string(messageBodyWithRetry))

				// Enviar el mensaje nuevamente a SQS
				if err := h.Utils.SendMessageToQueue(
					ctx, h.SQSClient, h.QueueURL, string(messageBodyWithRetry), messageID); err != nil {
					h.Logger.LogError("Error al reenviar el mensaje a SQS", err, messageID)
					return err
				}
			} else {
				h.Logger.LogError("Se alcanzó el máximo de reintentos", nil, messageID)
			}
		}
	}
	return nil
}

// printSQSEvent...
func printSQSEvent(sqsEvent events.SQSEvent) {
	eventJSON, err := json.MarshalIndent(sqsEvent, "", "  ")
	if err != nil {
		fmt.Printf("Error al convertir el evento a JSON: %v\n", err)
		return
	}
	fmt.Printf("\n--- Evento SQS recibido ---\n%s\n--- Fin del evento ---\n", string(eventJSON))
}
