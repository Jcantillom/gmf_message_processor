package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/service"
	"gmf_message_processor/internal/utils"
)

// SQSHandler es responsable de recibir y procesar mensajes desde SQS.
type SQSHandler struct {
	PlantillaService *service.PlantillaService
	SQSClient        *aws.SQSClient
}

// NewSQSHandler crea una nueva instancia de SQSHandler.
func NewSQSHandler(plantillaService *service.PlantillaService, sqsClient *aws.SQSClient) *SQSHandler {
	return &SQSHandler{
		PlantillaService: plantillaService,
		SQSClient:        sqsClient,
	}
}

// ProcessMessage procesa un solo mensaje desde la cola de SQS.
func (h *SQSHandler) ProcessMessage(ctx context.Context) error {

	// Recibir un mensaje de SQS
	output, err := h.SQSClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &h.SQSClient.QueueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     5,
	})
	if err != nil {
		logs.LogError("Error al recibir mensaje de SQS: %v", err)
		return fmt.Errorf("error general al recibir mensaje de SQS: %v", err)
	}

	if len(output.Messages) == 0 {
		logs.LogWarn("No hay mensajes en la cola de SQS.", "queue_url", h.SQSClient.QueueURL)
		return nil
	}

	// Procesar el primer mensaje en la cola
	message := output.Messages[0]

	// Extraer el cuerpo del mensaje
	messageBody, err := utils.ExtractMessageBody(*message.Body)
	if err != nil {
		logs.LogError("Error extrayendo el cuerpo del mensaje: %v", err)
		return fmt.Errorf("error al extraer el cuerpo del mensaje: %v", err)
	}

	// Validar el mensaje
	validMsg, err := utils.ValidateSQSMessage(messageBody)
	if err != nil {
		logs.LogError("Error validando el mensaje: %v", err)
		return fmt.Errorf("error al validar el mensaje: %v", err)
	}

	// Llamar al servicio para manejar la lógica de negocio
	if err := h.PlantillaService.HandlePlantilla(ctx, validMsg); err != nil {
		logs.LogError("Error en la lógica de negocio al procesar el mensaje: %v", err)
		return fmt.Errorf("error en la lógica de negocio al procesar el mensaje: %v", err)
	}

	// Si el mensaje se procesó correctamente, lo eliminamos de la cola
	err = utils.DeleteMessageFromQueue(ctx, h.SQSClient.Client, h.SQSClient.QueueURL, message.ReceiptHandle)
	if err != nil {
		logs.LogError("Error al eliminar el mensaje de la cola SQS: %v", err)
		return fmt.Errorf("error al eliminar el mensaje de la cola SQS: %v", err)
	}

	logs.LogInfo("Mensaje procesado correctamente")
	return nil
}
