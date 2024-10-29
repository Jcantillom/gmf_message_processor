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

func TestNonHTTPSURLInProd(t *testing.T) {
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	invalidURL := "http://non-https-url"
	err := validateQueueURL(invalidURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only HTTPS URLs are allowed for non-local environments")
}

// Test para verificar la creación del cliente SQS con una URL válida
func TestNewSQSClientValidURL(t *testing.T) {
	// Configurar APP_ENV como "local"
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Reiniciar Viper para asegurarse de que tome la variable de entorno
	viper.Reset()
	viper.AutomaticEnv()

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

func TestNewSQSClientProdEnv(t *testing.T) {
	// Configurar APP_ENV como "prod"
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	// Usar una URL HTTPS válida para entornos de producción
	validProdQueueURL := "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue"

	client, err := NewSQSClient(validProdQueueURL, mockLoadConfigFunc)

	// Verificar los resultados
	assert.NoError(t, err)                              // No debe haber error
	assert.NotNil(t, client)                            // El cliente no debe ser nil
	assert.Equal(t, validProdQueueURL, client.QueueURL) // La URL debe coincidir
}

func TestSQSClientGetQueueURL(t *testing.T) {
	client := &SQSClient{
		QueueURL: queueURL,
	}
	assert.Equal(t, queueURL, client.GetQueueURL())
}

// Test para verificar que DeleteMessage funciona correctamente con el mock
func TestSQSClientDeleteMessage(t *testing.T) {
	// Configura APP_ENV como "local"
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Reinicia Viper y asegúrate de que use las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Crear el mock del cliente SQS
	mockSQS := new(MockSQSClient)

	client := &SQSClient{
		Client:   mockSQS,
		QueueURL: queueURL,
	}

	mockInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(client.GetQueueURL()),
		ReceiptHandle: aws.String("test-receipt-handle"),
	}
	mockOutput := &sqs.DeleteMessageOutput{}

	// Configurar el mock para devolver el valor esperado
	mockSQS.On("DeleteMessage", mock.Anything, mockInput, mock.Anything).Return(mockOutput, nil)

	// Llamar a DeleteMessage
	result, err := client.DeleteMessage(context.TODO(), mockInput)

	// Verificar los resultados
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, mockOutput, result)

	// Verificar que se cumplieron las expectativas del mock
	mockSQS.AssertExpectations(t)
}

func TestSQSClientSendMessage(t *testing.T) {
	// Configura APP_ENV como "local"
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Reinicia Viper para asegurarse de que use las variables de entorno
	viper.Reset()
	viper.AutomaticEnv()

	// Crear el mock del cliente SQS
	mockSQS := new(MockSQSClient)

	// Crear el cliente SQS utilizando el mock
	client := &SQSClient{
		Client:   mockSQS,
		QueueURL: queueURL, // Debe coincidir exactamente con el mock
	}

	// Definir los inputs y outputs del mensaje
	mockInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(client.GetQueueURL()),
		MessageBody: aws.String("test message"),
	}
	mockOutput := &sqs.SendMessageOutput{
		MessageId: aws.String("12345"), // Simulando un ID de mensaje
	}

	// Configurar el mock para devolver el valor esperado
	mockSQS.On("SendMessage", mock.Anything, mockInput, mock.Anything).Return(mockOutput, nil)

	// Llamar a SendMessage
	result, err := client.SendMessage(context.TODO(), mockInput)

	// Verificar los resultados
	assert.NoError(t, err) // No debe haber error
	assert.NotNil(t, result)
	assert.Equal(t, "12345", *result.MessageId) // Verifica el ID del mensaje

	// Verificar que se cumplieron las expectativas del mock
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

func TestDeleteMessageInputValidationFailed(t *testing.T) {
	mockClient := new(MockSQSClient)
	client := &SQSClient{
		Client:   mockClient,
		QueueURL: queueURL,
	}

	// Usar un input inválido (input nil)
	_, err := client.DeleteMessage(context.TODO(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sanitization failed: input cannot be nil")
}

func TestDeleteMessageEmptyReceiptHandle(t *testing.T) {
	client := &SQSClient{
		QueueURL: queueURL,
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(""), // Receipt handle vacío
	}

	_, err := client.DeleteMessage(context.TODO(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "receipt handle is required and cannot be empty")
}

func TestInvalidQueueURL(t *testing.T) {
	invalidURL := "invalid-url"
	err := validateQueueURL(invalidURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid queue URL")
}

func TestSendMessageEmptyBody(t *testing.T) {
	client := &SQSClient{
		QueueURL: queueURL,
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(""), // Cuerpo del mensaje vacío
	}

	// Capturar ambos valores: el resultado y el error
	_, err := client.SendMessage(context.TODO(), input)

	// Verificar que el error no sea nulo y contenga el mensaje esperado
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "message body is required and cannot be empty")
}

func TestSendMessageEmptyQueueURL(t *testing.T) {
	client := &SQSClient{
		QueueURL: "",
	}

	input := &sqs.SendMessageInput{
		MessageBody: aws.String("test message"),
	}

	_, err := client.SendMessage(context.TODO(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue URL is required and cannot be empty")
}

func TestSendMessageInputNil(t *testing.T) {
	client := &SQSClient{
		QueueURL: queueURL,
	}

	// Llamar a SendMessage con input nil
	_, err := client.SendMessage(context.TODO(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input cannot be nil")
}

func TestDeleteMessageFailed(t *testing.T) {
	// Crear un mock del cliente SQS
	mockSQS := new(MockSQSClient)

	client := &SQSClient{
		Client:   mockSQS,
		QueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/my-queue", // URL válida
	}

	// Definir el input con datos válidos
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(client.GetQueueURL()),
		ReceiptHandle: aws.String("valid-receipt-handle"),
	}

	// Configurar el mock para que devuelva un error y un valor *sqs.DeleteMessageOutput nil
	mockSQS.On("DeleteMessage", mock.Anything, input, mock.Anything).
		Return((*sqs.DeleteMessageOutput)(nil), fmt.Errorf("failed to delete message"))

	// Llamar a DeleteMessage usando el cliente mockeado
	_, err := client.DeleteMessage(context.TODO(), input)

	// Verificar los resultados
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete message")
	mockSQS.AssertExpectations(t)
}

func TestDeleteMessageEmptyQueueURL(t *testing.T) {
	client := &SQSClient{
		QueueURL: "",
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(""),
		ReceiptHandle: aws.String("valid-handle"),
	}

	_, err := client.DeleteMessage(context.TODO(), input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue URL is required and cannot be empty")
}
