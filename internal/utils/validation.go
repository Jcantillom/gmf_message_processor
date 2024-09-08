package utils

import (
	"encoding/json"
	"errors"
	"gmf_message_processor/internal/models"
	"log"
)

// ValidateSQSMessage valida la estructura del mensaje JSON y verifica los campos obligatorios.
func ValidateSQSMessage(body string) (*models.SQSMessage, error) {
	var msg models.SQSMessage

	// Deserialización del mensaje JSON
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		return nil, errors.New("invalid JSON format")
	}

	// Validar que el campo obligatorio IdPlantilla esté presente
	if msg.IDPlantilla == "" {
		return nil, errors.New("IdPlantilla is required")
	}

	log.Println("Formato de mensaje válido 😉")

	return &msg, nil
}
