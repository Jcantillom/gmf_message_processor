package utils_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
)

// MockSQSAPI is a mock implementation of the SQSAPI interface.
type MockSQSAPI struct {
	mock.Mock
}

// DeleteMessage is a mock implementation of the DeleteMessage function.
func (m *MockSQSAPI) DeleteMessage(
	ctx context.Context,
	input *sqs.DeleteMessageInput,
	opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

// GetQueueURL is a mock implementation of the GetQueueURL function.
func (m *MockSQSAPI) GetQueueURL() string {
	return ""
}

// SendMessage is a mock implementation of the SendMessage function.
func (m *MockSQSAPI) SendMessage(
	ctx context.Context,
	input *sqs.SendMessageInput,
	opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

// TestExtractMessageBody tests the ExtractMessageBody function.
func TestExtractMessageBody(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	validMessage := `{"key": "value"}`
	invalidMessage := `invalid JSON`

	// Test valid message
	body, err := u.ExtractMessageBody(validMessage, messageID)
	assert.NoError(t, err)
	assert.Equal(t, validMessage, body)

	// Test invalid message
	_, err = u.ExtractMessageBody(invalidMessage, messageID)
	assert.Error(t, err)
	assert.Equal(t, "error deserializando el mensaje de SQS", err.Error())
}

// TestDeleteMessageFromQueue tests the DeleteMessageFromQueue function.
func TestDeleteMessageFromQueue(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	mockSQS := new(MockSQSAPI)
	queueURL := "https://example.com/queue"
	receiptHandle := "testReceiptHandle"

	mockSQS.On("DeleteMessage", mock.Anything, mock.Anything).Return(&sqs.DeleteMessageOutput{}, nil)

	err := u.DeleteMessageFromQueue(
		context.TODO(),
		mockSQS,
		queueURL,
		&receiptHandle,
		messageID,
	)
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

// TestValidateSQSMessage tests the ValidateSQSMessage function.
func TestValidateSQSMessage(t *testing.T) {
	u := &utils.Utils{}

	validMessage := models.SQSMessage{IDPlantilla: "123"}
	validBody, _ := json.Marshal(validMessage)
	invalidMessage := `{"id_plantilla": ""}`

	// Test valid message
	msg, err := u.ValidateSQSMessage(string(validBody))
	assert.NoError(t, err)
	assert.NotNil(t, msg)

	// Test invalid message (missing IDPlantilla)
	_, err = u.ValidateSQSMessage(invalidMessage)
	assert.Error(t, err)
	assert.Equal(t, "id_plantilla is required", err.Error())
}

// TestSendMessageToQueue tests the SendMessageToQueue function.
func TestSendMessageToQueue(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	mockSQS := new(MockSQSAPI)
	queueURL := "https://example.com/queue"
	messageBody := "test message"

	// Set SQS_MESSAGE_DELAY env variable
	os.Setenv("SQS_MESSAGE_DELAY", "5")
	defer os.Unsetenv("SQS_MESSAGE_DELAY")

	mockSQS.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, nil)

	err := u.SendMessageToQueue(
		context.TODO(),
		mockSQS,
		queueURL,
		messageBody,
		messageID,
	)
	assert.NoError(t, err)
	mockSQS.AssertExpectations(t)
}

// TestReplacePlaceholders tests the ReplacePlaceholders function.
func TestReplacePlaceholders(t *testing.T) {
	text := "Hello, {name}!"
	params := map[string]string{
		"{name}": "Juan",
	}

	result := utils.ReplacePlaceholders(text, params)
	assert.Equal(t, "Hello, Juan!", result)
}

// TestGetMaxRetries tests the GetMaxRetries function.
func TestGetMaxRetries(t *testing.T) {
	// Test with no env variable set
	os.Unsetenv("MAX_RETRIES")
	assert.Equal(t, 3, utils.GetMaxRetries())

	// Test with valid env variable
	os.Setenv("MAX_RETRIES", "5")
	defer os.Unsetenv("MAX_RETRIES")
	assert.Equal(t, 5, utils.GetMaxRetries())

	// Test with invalid env variable
	os.Setenv("MAX_RETRIES", "invalid")
	assert.Equal(t, 3, utils.GetMaxRetries())
}

func TestValidateSQSMessage_InvalidJSON(t *testing.T) {
	u := &utils.Utils{}

	invalidMessage := `invalid json`

	// Test invalid JSON format
	_, err := u.ValidateSQSMessage(invalidMessage)
	assert.Error(t, err)
	assert.Equal(t, "invalid JSON format", err.Error())
}

func TestSendMessageToQueue_InvalidDelayEnvVar(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	mockSQS := new(MockSQSAPI)
	queueURL := "https://example.com/queue"
	messageBody := "test message"

	// Set invalid SQS_MESSAGE_DELAY env variable
	os.Setenv("SQS_MESSAGE_DELAY", "invalid")
	defer os.Unsetenv("SQS_MESSAGE_DELAY")

	// Ejecutar la función que debería fallar debido al valor inválido de la variable de entorno
	err := u.SendMessageToQueue(
		context.TODO(),
		mockSQS,
		queueURL,
		messageBody,
		messageID,
	)

	// Verificar que ocurrió un error
	assert.Error(t, err)
	// Verificar que el error contiene el mensaje específico de strconv.Atoi
	assert.Contains(t, err.Error(), "invalid syntax") // Verifica que el error contenga el mensaje de strconv.Atoi
}

func TestSendMessageToQueue_Failure(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	mockSQS := new(MockSQSAPI)
	queueURL := "https://example.com/queue"
	messageBody := "test message"

	// Set SQS_MESSAGE_DELAY env variable
	os.Setenv("SQS_MESSAGE_DELAY", "5")
	defer os.Unsetenv("SQS_MESSAGE_DELAY")

	// Simular que SQS devuelve un error pero sigue devolviendo un *sqs.SendMessageOutput válido
	mockSQS.On("SendMessage", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, errors.New("SQS error"))

	err := u.SendMessageToQueue(
		context.TODO(),
		mockSQS,
		queueURL,
		messageBody,
		messageID,
	)

	// Verificar que ocurrió un error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SQS error")
}

func TestDeleteMessageFromQueue_Failure(t *testing.T) {
	u := &utils.Utils{}
	messageID := "testMessageID"
	mockSQS := new(MockSQSAPI)
	queueURL := "https://example.com/queue"
	receiptHandle := "testReceiptHandle"

	// Simular que SQS devuelve un error pero aún devuelve un *sqs.DeleteMessageOutput válido
	mockSQS.On("DeleteMessage", mock.Anything, mock.Anything).Return(&sqs.DeleteMessageOutput{}, errors.New("SQS delete error"))

	err := u.DeleteMessageFromQueue(
		context.TODO(),
		mockSQS,
		queueURL,
		&receiptHandle,
		messageID,
	)

	// Verificar que ocurrió un error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SQS delete error")
}
