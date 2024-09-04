package processor

import (
	"encoding/json"
	"fmt"
)

// Message representa la estructura esperada del mensaje JSON.
type Message struct {
	BucketName           string `json:"bucket_name"`
	FolderName           string `json:"folder_name"`
	FileName             string `json:"file_name"`
	FileID               int64  `json:"file_id"`
	ResponseProcessingID int    `json:"response_processing_id"`
}

// ValidateMessage valida el formato del mensaje recibido.
func ValidateMessage(body string) (*Message, error) {
	var msg Message
	err := json.Unmarshal([]byte(body), &msg)
	if err != nil {
		return nil, fmt.Errorf("formato de mensaje no válido 🚫 %v", err)
	}

	// Validar que los campos requeridos no estén vacíos o tengan valores incorrectos.
	if msg.BucketName == "" || msg.FolderName == "" || msg.FileName == "" || msg.FileID == 0 || msg.ResponseProcessingID == 0 {
		return nil, fmt.Errorf("mensaje JSON tiene campos vacíos o inválidos 🚫")
	}

	return &msg, nil
}
