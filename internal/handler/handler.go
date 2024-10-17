package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
	"strings"
)

// Interfaces para inyección de dependencias
type SQSHandlerInterface interface {
	HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error
}

type LogInterface interface {
	LogError(message string, err error, messageID string)
	LogInfo(message, messageID string)
}

type PlantillaServiceInterface interface {
	HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error
}

type UtilsInterface interface {
	ExtractMessageBody(body, messageID string) (string, error)
	ValidateSQSMessage(messageBody string) (*models.SQSMessage, error)
	DeleteMessageFromQueue(ctx context.Context, client aws.SQSAPI, queueURL string, receiptHandle *string, messageID string) error
	SendMessageToQueue(ctx context.Context, client aws.SQSAPI, queueURL, messageBody, messageID string) error
}

// Estructura principal del manejador
type SQSHandler struct {
	PlantillaService PlantillaServiceInterface
	SQSClient        aws.SQSAPI
	Utils            UtilsInterface
	Logger           LogInterface
	QueueURL         string
}

// Constructor del manejador
func NewSQSHandler(
	plantillaService PlantillaServiceInterface,
	sqsClient aws.SQSAPI,
	utils UtilsInterface,
	logger LogInterface,
	queueURL string,
) *SQSHandler {
	return &SQSHandler{
		PlantillaService: plantillaService,
		SQSClient:        sqsClient,
		Utils:            utils,
		Logger:           logger,
		QueueURL:         queueURL,
	}
}

// Función principal que maneja el evento Lambda
func (h *SQSHandler) HandleLambdaEvent(ctx context.Context, sqsEvent events.SQSEvent) error {
	//imprime el evento
	printSQSEvent(sqsEvent)
	for _, record := range sqsEvent.Records {
		messageID := record.MessageId

		if err := h.processMessage(ctx, record, messageID); err != nil {
			h.Logger.LogError("Error procesando el mensaje", err, messageID)
			return err
		}
	}
	return nil
}

// Procesa un mensaje individual
func (h *SQSHandler) processMessage(ctx context.Context, record events.SQSMessage, messageID string) error {
	// Elimina el mensaje de la cola inmediatamente antes de iniciar el procesamiento
	if err := h.Utils.DeleteMessageFromQueue(
		ctx, h.SQSClient, h.QueueURL, &record.ReceiptHandle, messageID); err != nil {
		return fmt.Errorf("Error eliminando mensaje de SQS: %w", err)
	}

	messageBody, err := h.Utils.ExtractMessageBody(record.Body, messageID)
	if err != nil {
		return fmt.Errorf("Error extrayendo cuerpo del mensaje: %w", err)
	}

	validMsg, err := h.Utils.ValidateSQSMessage(messageBody)
	if err != nil {
		return fmt.Errorf("Error validando mensaje: %w", err)
	}

	defer h.handleRecovery(validMsg, messageID)

	if err := h.PlantillaService.HandlePlantilla(ctx, validMsg, messageID); err != nil {
		return h.retryMessage(ctx, validMsg, messageID, err)
	}
	return nil
}

// Maneja la recuperación en caso de panic
func (h *SQSHandler) handleRecovery(validMsg *models.SQSMessage, messageID string) {
	if r := recover(); r != nil {
		h.Logger.LogError("Error al procesar el mensaje", fmt.Errorf("%v", r), messageID)
		h.retryMessage(context.Background(), validMsg, messageID, nil)
	}
}

var jsonMarshal = json.Marshal

// Reintenta el envío de un mensaje a SQS
func (h *SQSHandler) retryMessage(ctx context.Context, msg *models.SQSMessage, messageID string, err error) error {
	if err != nil {
		h.Logger.LogError("Error al procesar el mensaje", err, messageID)
	}

	msg.RetryCount++
	if msg.RetryCount > utils.GetMaxRetries() {
		h.Logger.LogError("Se alcanzó el máximo de reintentos", nil, messageID)
		return nil
	}

	h.Logger.LogInfo(fmt.Sprintf("Reintentando el mensaje. Conteo actual: %d", msg.RetryCount), messageID)

	messageBodyWithRetry, err := jsonMarshal(msg)
	if err != nil {
		return fmt.Errorf("Error convirtiendo mensaje a JSON: %w", err)
	}

	if err := h.Utils.SendMessageToQueue(ctx, h.SQSClient, h.QueueURL, string(messageBodyWithRetry), messageID); err != nil {
		return fmt.Errorf("Error reenviando mensaje a SQS: %w", err)
	}
	return nil
}

// En handler.go
var jsonMarshalIndent = json.MarshalIndent
var logDebug = logs.LogDebug

func printSQSEvent(sqsEvent events.SQSEvent) {
	eventJSON, err := jsonMarshalIndent(sqsEvent, "", "  ") // Usamos el wrapper aquí
	if err != nil {
		logDebug(fmt.Sprintf("Error al convertir el evento a JSON: %v", err), "")
		return
	}

	singleLineJSON := strings.ReplaceAll(string(eventJSON), "\n", " ")
	logDebug(
		fmt.Sprintf("--- Evento SQS --- %s --- Fin del Evento SQS ---", singleLineJSON), "")
}
