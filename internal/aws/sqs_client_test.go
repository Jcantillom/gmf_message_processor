package aws

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock de la función de carga de configuración
func mockLoadConfigFunc(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return aws.Config{}, nil
}

// Mock para el cliente SQS
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, input *sqs.DeleteMessageInput, opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) GetQueueURL() string {
	args := m.Called()
	return args.String(0)
}

// Test para verificar la creación del cliente SQS con una URL válida
func TestNewSQSClient_ValidURL(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", mockLoadConfigFunc)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)
}

// Test para verificar el manejo de una URL inválida
func TestNewSQSClient_InvalidURL(t *testing.T) {
	_, err := NewSQSClient("invalid-url", mockLoadConfigFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid queue URL")
}

// Test para manejar un error al cargar la configuración
func TestNewSQSClient_LoadConfigError(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	mockLoadConfigFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, fmt.Errorf("unable to load AWS SDK config")
	}

	_, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", mockLoadConfigFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load AWS SDK config")
}

// Test para verificar la creación de un cliente SQS en un entorno de producción
func TestNewSQSClient_ProdEnv(t *testing.T) {
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", mockLoadConfigFunc)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", client.QueueURL)
}
