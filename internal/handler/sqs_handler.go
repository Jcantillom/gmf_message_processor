package handler

import (
	"context"
	"gmf_message_processor/internal/service"
	"gmf_message_processor/internal/utils"
	"log"
)

// ProcessSQSMessages procesa los mensajes desde un archivo JSON en la carpeta test_data.
func ProcessSQSMessages(ctx context.Context) (int, error) {
	// Leer el evento SQS desde el archivo JSON usando la función ReadSQSEventFromFile de utils.
	sqsEvent, err := utils.ReadSQSEventFromFile("test_data/message.json")
	if err != nil {
		return 0, err
	}

	processed := 0 // Contador de mensajes procesados

	// Procesar cada mensaje en el evento
	for _, message := range sqsEvent.Records {
		// Validar el mensaje
		validMsg, err := utils.ValidateSQSMessage(message.Body)
		if err != nil {
			log.Printf("Error validando el mensaje ❌: %v", err)
			continue
		}

		// Llamar al servicio para manejar la lógica de negocio
		if err := service.HandlePlantilla(ctx, validMsg); err != nil {
			log.Printf("Error al procesar el mensaje para IDPlantilla: %s: %v", validMsg.IDPlantilla, err)
			return processed, err // Devolver el error inmediatamente al encontrar un problema
		} else {
			log.Printf("Mensaje procesado con éxito para IDPlantilla ✅: %s", validMsg.IDPlantilla)
			processed++ // Incrementar el contador de mensajes procesados
		}
	}

	return processed, nil
}
