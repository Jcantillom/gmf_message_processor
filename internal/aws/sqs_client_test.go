package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/spf13/viper"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/stretchr/testify/assert"
)

func defaultLoadConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, optFns...)
}

func TestNewSQSClient_ValidURL(t *testing.T) {
	// Establecer la variable de entorno
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Asegurarse de que Viper vuelva a cargar las variables de entorno
	viper.AutomaticEnv()

	// Ahora ejecuta la prueba
	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", defaultLoadConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)
}

func TestNewSQSClient_InvalidURL(t *testing.T) {
	_, err := NewSQSClient("invalid-url", defaultLoadConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid queue URL")
}

func TestNewSQSClient_UnknownAppEnv(t *testing.T) {
	os.Setenv("APP_ENV", "unknown")
	defer os.Unsetenv("APP_ENV")

	_, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", defaultLoadConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown APP_ENV")
}

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

func TestNewSQSClient_ProdEnv(t *testing.T) {
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", defaultLoadConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", client.QueueURL)
}

func TestNewSQSClient_LocalEnv(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", defaultLoadConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)
}

func TestNewSQSClient_QAEnv(t *testing.T) {
	os.Setenv("APP_ENV", "qa")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", defaultLoadConfig)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", client.QueueURL)
}
func TestNewSQSClient_LocalStack_ValidEndpoint(t *testing.T) {
	// Establecer el entorno como "local"
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Crear un nuevo cliente SQS utilizando la función de carga de configuración predeterminada
	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", defaultLoadConfig)

	// Asegurarse de que no haya errores
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)
}

func TestNewSQSClient_DevEnv(t *testing.T) {
	os.Setenv("APP_ENV", "dev")
	defer os.Unsetenv("APP_ENV")

	// Usar una URL válida
	client, err := NewSQSClient("https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", defaultLoadConfig)

	// Verificar que no hay error
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewSQSClient_AWSInvalidService(t *testing.T) {
	os.Setenv("APP_ENV", "prod")
	defer os.Unsetenv("APP_ENV")

	// Mock para el resolvedor de endpoints que devuelve un EndpointNotFoundError
	mockEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Función de carga de configuración que utiliza el resolvedor de endpoints mock
	mockLoadConfigFunc := func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return config.LoadDefaultConfig(ctx, append(optFns, config.WithEndpointResolver(mockEndpointResolver))...)
	}

	// Crear un nuevo cliente SQS
	client, err := NewSQSClient("https://sqs.us-east-1.amazonaws.com/000000000000/my-queue", mockLoadConfigFunc)

	// Verificar que no hay error

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestNewSQSClient_LocalStack_UnknownEndpoint(t *testing.T) {
	// Establecer el entorno como "local"
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Mock para el resolvedor de endpoints que devuelve un error
	mockEndpointResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == sqs.ServiceID && region == "us-east-1" {
			return aws.Endpoint{
				URL:           "http://localhost:4566",
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})

	// Función de carga de configuración que utiliza el resolvedor de endpoints mock
	mockLoadConfigFunc := func(
		ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return config.LoadDefaultConfig(ctx, append(optFns, config.WithEndpointResolver(mockEndpointResolver))...)
	}

	// Crear un nuevo cliente SQS con una URL válida
	client, err := NewSQSClient(
		"http://localhost:4566/000000000000/my-queue", mockLoadConfigFunc)

	// Verificar que no hay error
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verificar que la URL de la cola es la de LocalStack
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)

	// Verificar que el resolvedor de endpoints devolvió un error al simular la llamada
	_, err = mockEndpointResolver("unknown-service", "us-east-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

	// Verificar que el resolvedor de endpoints devolvió un error al simular la llamada
	_, err = mockEndpointResolver("sqs", "us-west-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

	// Verificar que el resolvedor de endpoints devolvió un error al simular la llamada
	_, err = mockEndpointResolver("sqs", "us-east-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

}

// Mock de un endpoint resolver
func mockEndpointResolver(service, region string) (aws.Endpoint, error) {
	if service == sqs.ServiceID && region == "us-east-1" {
		return aws.Endpoint{
			URL:           "http://localhost:4566",
			SigningRegion: "us-east-1",
		}, nil
	}
	return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
}

// Test para verificar el comportamiento del resolvedor
func TestNewSQSClient_EndpointResolver(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	// Función de carga de configuración que utiliza el resolvedor de endpoints mock
	mockLoadConfigFunc := func(
		ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return config.LoadDefaultConfig(
			ctx, append(optFns, config.WithEndpointResolver(aws.EndpointResolverFunc(mockEndpointResolver)))...)
	}

	// Crear un nuevo cliente SQS
	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", mockLoadConfigFunc)

	// Verificar que no hay error y el cliente es válido
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Verificar que la URL de la cola es la correcta
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)

	// Probar el caso donde el resolvedor de endpoints debería fallar
	_, err = mockEndpointResolver("invalid-service", "us-east-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

	// Probar el caso donde el resolvedor de endpoints debería fallar
	_, err = mockEndpointResolver("sqs", "us-west-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

	// Probar el caso donde el resolvedor de endpoints debería fallar
	_, err = mockEndpointResolver("sqs", "us-east-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown endpoint requested")

}
func mockLoadConfigFunc(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, append(optFns, config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service == sqs.ServiceID && region == "us-east-1" {
			return aws.Endpoint{
				URL:           "http://localhost:4566", // URL de LocalStack
				SigningRegion: "us-east-1",
			}, nil
		}
		return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
	})))...)

}

func TestCase1(t *testing.T) {
	os.Setenv("APP_ENV", "local")
	defer os.Unsetenv("APP_ENV")

	client, err := NewSQSClient("http://localhost:4566/000000000000/my-queue", mockLoadConfigFunc)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:4566/000000000000/my-queue", client.QueueURL)
}
func TestGetEndpointResolver_LocalEnv(t *testing.T) {
	// Configurar el valor de APP_ENV como "local"
	viper.Set("APP_ENV", "local")

	resolver, err := getEndpointResolver()
	if err != nil {
		t.Fatalf("No se esperaba error al obtener el resolvedor de endpoints: %v", err)
	}

	// Probar el caso cuando el servicio es SQS y la región es us-east-1
	endpoint, err := resolver.ResolveEndpoint(sqs.ServiceID, "us-east-1")
	if err != nil {
		t.Fatalf("No se esperaba error al resolver el endpoint: %v", err)
	}

	expectedURL := "http://localhost:4566"
	if endpoint.URL != expectedURL {
		t.Errorf("Se esperaba el URL del endpoint: %s, pero se obtuvo: %s", expectedURL, endpoint.URL)
	}

	expectedRegion := "us-east-1"
	if endpoint.SigningRegion != expectedRegion {
		t.Errorf("Se esperaba la región de firma: %s, pero se obtuvo: %s", expectedRegion, endpoint.SigningRegion)
	}
}

func TestGetEndpointResolver_UnknownService(t *testing.T) {
	// Configurar el valor de APP_ENV como "local"
	viper.Set("APP_ENV", "local")

	resolver, err := getEndpointResolver()
	if err != nil {
		t.Fatalf("No se esperaba error al obtener el resolvedor de endpoints: %v", err)
	}

	// Probar el caso cuando el servicio es desconocido
	_, err = resolver.ResolveEndpoint("unknown_service", "us-east-1")
	if err == nil {
		t.Error("Se esperaba un error para un servicio desconocido, pero no se obtuvo ningún error")
	}

	expectedErr := "unknown endpoint requested"
	if err.Error() != expectedErr {
		t.Errorf("Se esperaba el error: %s, pero se obtuvo: %s", expectedErr, err.Error())
	}
}

func TestGetEndpointResolver_UnknownRegion(t *testing.T) {
	// Configurar el valor de APP_ENV como "local"
	viper.Set("APP_ENV", "local")

	resolver, err := getEndpointResolver()
	if err != nil {
		t.Fatalf("No se esperaba error al obtener el resolvedor de endpoints: %v", err)
	}

	// Probar el caso cuando la región es desconocida
	_, err = resolver.ResolveEndpoint(sqs.ServiceID, "unknown-region")
	if err == nil {
		t.Error("Se esperaba un error para una región desconocida, pero no se obtuvo ningún error")
	}

	expectedErr := "unknown endpoint requested"
	if err.Error() != expectedErr {
		t.Errorf("Se esperaba el error: %s, pero se obtuvo: %s", expectedErr, err.Error())
	}
}

func TestGetEndpointResolver_UnknownAppEnv(t *testing.T) {
	// Configurar el valor de APP_ENV como "desconocido"
	viper.Set("APP_ENV", "unknown")

	_, err := getEndpointResolver()
	if err == nil {
		t.Fatal("Se esperaba un error para un APP_ENV desconocido, pero no se obtuvo ningún error")
	}

	expectedErr := "unknown APP_ENV: unknown"
	if err.Error() != expectedErr {
		t.Errorf("Se esperaba el error: %s, pero se obtuvo: %s", expectedErr, err.Error())
	}
}

func TestGetEndpointResolver_EndpointNotFoundError(t *testing.T) {
	// Configurar el valor de APP_ENV como "dev"
	viper.Set("APP_ENV", "dev")

	resolver, err := getEndpointResolver()
	if err != nil {
		t.Fatalf("No se esperaba error al obtener el resolvedor de endpoints: %v", err)
	}

	// Probar el caso cuando se devuelve aws.EndpointNotFoundError
	_, err = resolver.ResolveEndpoint(sqs.ServiceID, "us-east-1")
	if err == nil {
		t.Error("Se esperaba un error aws.EndpointNotFoundError, pero no se obtuvo ningún error")
	}

	// Verificar que el error sea de tipo aws.EndpointNotFoundError
	if _, ok := err.(*aws.EndpointNotFoundError); !ok {
		t.Errorf("Se esperaba un error de tipo aws.EndpointNotFoundError, pero se obtuvo: %T", err)
	}
}
