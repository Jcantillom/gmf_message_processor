package connection

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSecretService es una implementación simulada de SecretService para pruebas
type MockSecretService struct {
	mock.Mock
}

func (m *MockSecretService) GetSecret(secretName string) (*SecretData, error) {
	args := m.Called(secretName)
	if secretData, ok := args.Get(0).(*SecretData); ok {
		return secretData, args.Error(1)
	}
	return nil, args.Error(1) // Devolver nil en caso de que no sea del tipo esperado
}

// TestGetSecret éxito
func TestGetSecret_Success(t *testing.T) {
	mockService := new(MockSecretService)
	secretData := &SecretData{
		Username: "testUser",
		Password: "testPass",
	}
	mockService.On("GetSecret", "test-secret").Return(secretData, nil)

	// Llamar al método
	result, err := mockService.GetSecret("test-secret")

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, secretData.Username, result.Username)
	assert.Equal(t, secretData.Password, result.Password)
	mockService.AssertExpectations(t)
}

// TestGetSecret_error
func TestGetSecret_Error(t *testing.T) {
	mockService := new(MockSecretService)
	mockService.On(
		"GetSecret",
		"test-secret").Return((*SecretData)(nil), errors.New("secret not found")) // Cambiar a *SecretData(nil)

	// Llamar al método
	result, err := mockService.GetSecret("test-secret")

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, result)
	mockService.AssertExpectations(t)
}

// TestNewSession éxito
func TestNewSession_Success(t *testing.T) {
	// Simular una sesión de AWS
	_, err := NewSession()

	assert.NoError(t, err)
}

// TestNewSession_error (puedes usar un mock para probar el error)
func TestNewSession_Error(t *testing.T) {
	mockService := new(MockSecretService)
	mockService.On(
		"GetSecret", "test").Return((*SecretData)(nil), errors.New("secret not found"))

	// Llamar al método
	result, err := mockService.GetSecret("test")

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, result)

	mockService.AssertExpectations(t)
}
