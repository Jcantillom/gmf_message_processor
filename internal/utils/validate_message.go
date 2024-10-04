package utils

import (
	"context"
	"encoding/json"
	"errors"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
)

func ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	var msg models.SQSMessage

	// Deserialization del mensaje JSON
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, errors.New("invalid JSON format")
	}

	// Validar que el campo obligatorio IdPlantilla esté presente
	if msg.IDPlantilla == "" {
		return nil, errors.New("IdPlantilla is required")
	}
	ctx := context.TODO()

	logs.LogFormatoMensajeValido(ctx)

	return &msg, nil
}
