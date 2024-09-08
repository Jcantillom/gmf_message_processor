package unit

import (
	"gmf_message_processor/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSQSMessage(t *testing.T) {
	validMessage := `{"IdPlantilla": "123", "parametro": ["test"]}`
	invalidMessage := `{"parametro": ["test"]}`

	// Test valid message
	_, err := utils.ValidateSQSMessage(validMessage)
	assert.Nil(t, err)

	// Test invalid message
	_, err = utils.ValidateSQSMessage(invalidMessage)
	assert.NotNil(t, err)
	assert.Equal(t, "IdPlantilla is required", err.Error())
}
