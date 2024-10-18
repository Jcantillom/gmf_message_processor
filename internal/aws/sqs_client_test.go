package aws

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock de la función de carga de configuración
func mockLoadConfigFunc(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return aws.Config{}, nil
}

// / Mock para el cliente SQS que implementa la interfaz SQSAPI
type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) DeleteMessage(
	ctx context.Context,
	input *sqs.DeleteMessageInput,
	opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) SendMessage(
	ctx context.Context,
	input *sqs.SendMessageInput,
	opts ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *MockSQSClient) GetQueueURL() string {
	args := m.Called()
	return args.String(0)
}

const queueURL = "http://localhost:4566/000000000000/my-queue"
const region = "us-east-1"

// Test para verificar la creación del cliente SQS con una URL válida
func TestNewSQSClientValidURL(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient(queueURL, mockLoadConfigFunc)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, queueURL, client.QueueURL)
}

// Test para verificar el manejo de una URL inválida
func TestNewSQSClientInvalidURL(t *testing.T) {
	_, err := NewSQSClient("invalid-url", mockLoadConfigFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid queue URL")
}

// Test para manejar un error al cargar la configuración
func TestNewSQSClientLoadConfigError(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	mockLoadConfigFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, fmt.Errorf("unable to load AWS SDK config")
	}

	_, err := NewSQSClient(queueURL, mockLoadConfigFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to load AWS SDK config")
}

// Test para verificar la creación de un cliente SQS en un entorno de producción
func TestNewSQSClientProdEnv(t *testing.T) {
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient(queueURL, mockLoadConfigFunc)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, queueURL, client.QueueURL)
}

func TestSQSClientGetQueueURL(t *testing.T) {
	client := &SQSClient{
		QueueURL: queueURL,
	}
	assert.Equal(t, queueURL, client.GetQueueURL())
}

// Test para verificar que DeleteMessage funciona correctamente con el mock
func TestSQSClientDeleteMessage(t *testing.T) {
	// Crear el mock del cliente SQS
	mockSQS := new(MockSQSClient)

	// Crear el cliente SQS utilizando el mock
	client := &SQSClient{
		Client:   mockSQS, // Usa el mock en lugar de *sqs.Client
		QueueURL: queueURL,
	}

	// Definir los inputs y outputs mockeados
	mockInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(client.GetQueueURL()),
		ReceiptHandle: aws.String("test-receipt-handle"),
	}
	mockOutput := &sqs.DeleteMessageOutput{}

	// Configurar el mock para que devuelva el valor esperado
	mockSQS.On("DeleteMessage", mock.Anything, mockInput, mock.Anything).Return(mockOutput, nil)

	// Llamar a DeleteMessage usando el cliente mockeado
	result, err := client.DeleteMessage(context.TODO(), mockInput)

	// Verificar los resultados
	assert.NoError(t, err)
	assert.Equal(t, mockOutput, result)
	mockSQS.AssertExpectations(t)
}

func TestSQSClientSendMessage(t *testing.T) {
	// Crear un mock del cliente SQS
	mockSQS := new(MockSQSClient)
	client := &SQSClient{
		Client:   mockSQS,
		QueueURL: queueURL,
	}

	// Definir los inputs y outputs mockeados
	mockInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(client.GetQueueURL()),
		MessageBody: aws.String("test message"),
	}
	mockOutput := &sqs.SendMessageOutput{}

	// Configurar el mock para que devuelva el valor esperado
	mockSQS.On("SendMessage", mock.Anything, mockInput, mock.Anything).Return(mockOutput, nil)

	// Llamar a SendMessage usando el cliente mockeado
	result, err := client.SendMessage(context.TODO(), mockInput)

	// Verificar los resultados
	assert.NoError(t, err)
	assert.Equal(t, mockOutput, result)
	mockSQS.AssertExpectations(t)
}

func TestGetEndpointResolverLocalEnv(t *testing.T) {
	// Configurar el entorno como local
	viper.Set("APP_ENV", "local")

	// Obtener el resolver
	resolver := getEndpointResolver()

	// Probar que devuelve el endpoint correcto para el servicio SQS en local
	endpoint, err := resolver.ResolveEndpoint(sqs.ServiceID, region)
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:4566", endpoint.URL)
	assert.Equal(t, region, endpoint.SigningRegion)
}

func TestGetEndpointResolverProdEnv(t *testing.T) {
	// Configurar el entorno como prod
	viper.Set("APP_ENV", "prod")

	// Obtener el resolver
	resolver := getEndpointResolver()

	// Probar que devuelve un error para entornos prod
	_, err := resolver.ResolveEndpoint(sqs.ServiceID, region)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested for service")
}

func TestGetEndpointResolverUnknownAppEnv(t *testing.T) {
	// Configurar el entorno con un valor desconocido
	viper.Set("APP_ENV", "staging")

	// Obtener el resolver
	resolver := getEndpointResolver()

	// Probar que devuelve un error para APP_ENV desconocido
	_, err := resolver.ResolveEndpoint(sqs.ServiceID, region)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown APP_ENV")
}

func TestGetEndpointResolverUnknownService(t *testing.T) {
	// Configurar el entorno como local
	viper.Set("APP_ENV", "local")

	// Obtener el resolver
	resolver := getEndpointResolver()

	// Probar que devuelve un error para un servicio desconocido
	_, err := resolver.ResolveEndpoint("unknown-service", "us-east-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown service")
}
