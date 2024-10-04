package handler

import (
	"context"
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
	logs.LogProcesandoMensajeSQS(ctx)

	// Recibir un mensaje de SQS
	output, err := h.SQSClient.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &h.SQSClient.QueueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     5,
	})
	if err != nil {
		logs.LogError(ctx, "Error al recibir mensaje de SQS: %v", err)
		return err
	}

	if len(output.Messages) == 0 {
		logs.LogWarn(ctx, "No hay mensajes en la cola de SQS.")
		return nil
	}

	// Procesar el primer mensaje en la cola
	message := output.Messages[0]

	// Extraer el cuerpo del mensaje
	messageBody, err := utils.ExtractMessageBody(*message.Body)
	if err != nil {
		logs.LogError(ctx, "Error extrayendo el cuerpo del mensaje: %v", err)
		return err
	}

	// Validar el mensaje
	validMsg, err := utils.ValidateSQSMessage(messageBody)
	if err != nil {
		logs.LogError(ctx, "Error validando el mensaje: %v", err)
		return err
	}

	// Llamar al servicio para manejar la lógica de negocio
	if err := h.PlantillaService.HandlePlantilla(ctx, validMsg); err != nil {
		logs.LogError(ctx, "Error al procesar el mensaje: %v", err)
		return err
	}

	logs.LogPlantillaEncontrada(ctx, validMsg.IDPlantilla)

	// Eliminar el mensaje de la cola una vez procesado
	err = utils.DeleteMessageFromQueue(ctx, h.SQSClient.Client, h.SQSClient.QueueURL, message.ReceiptHandle)
	if err != nil {
		return err
	}
	return nil
}
