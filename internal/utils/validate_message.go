package utils

import (
	"encoding/json"
	"errors"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
)

func ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	var msg models.SQSMessage

	// Deserializar el cuerpo JSON al modelo SQSMessage
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, errors.New("invalid JSON format")
	}

	// Validar que el campo obligatorio id_plantilla esté presente
	if msg.IDPlantilla == "" {
		return nil, errors.New("id_plantilla is required")
	}

	logs.LogWarn("Mensaje SQS válido", "id_plantilla", msg.IDPlantilla)

	return &msg, nil
}
