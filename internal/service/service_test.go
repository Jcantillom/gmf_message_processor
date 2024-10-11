package service

import (
	"context"
	"github.com/stretchr/testify/assert"
	"gmf_message_processor/internal/models"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockPlantillaRepository struct {
	mock.Mock
}

func (m *MockPlantillaRepository) CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error) {
	args := m.Called(idPlantilla)
	return args.Bool(0), args.Get(1).(*models.Plantilla), args.Error(2)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(
	remitente,
	destinatarios,
	asunto,
	cuerpo,
	messageID string) error {
	args := m.Called(remitente, destinatarios, asunto, cuerpo)
	return args.Error(0)
}

func TestHandlePlantilla_InvalidParameters(t *testing.T) {
	// Mock del repositorio para que devuelva una plantilla válida
	repo := new(MockPlantillaRepository)
	emailService := new(MockEmailService)

	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Cuerpo de prueba",
		Remitente:    "test@example.com",
		Destinatario: "dest@example.com",
		Adjunto:      false,
	}, nil)

	// Mock del servicio de email para que devuelva un error
	emailService.On(
		"SendEmail",
		"test@example.com",
		"dest@example.com", "Asunto de prueba",
		"Cuerpo de prueba").Return(nil)

	// crear una nueva instancia de Service
	service := NewPlantillaService(repo, emailService)

	// Llamar a HandlePlantilla con un mensaje que no tiene parámetros
	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)
	assert.NoError(t, err, "No debería haber un error cuando no se proporcionan parámetros")

	// Verificar que el repositorio y el servicio de correo fueron invocados correctamente
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)

}

func TestHandlePlantilla_PlantillaNotFound(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(false, (*models.Plantilla)(nil), nil)

	service := NewPlantillaService(repo, nil)

	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)
	assert.Error(t, err, "Debería haber un error cuando la plantilla no existe en la base de datos")
	repo.AssertExpectations(t)
}

func TestHandlePlantilla_Success(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Cuerpo de prueba",
		Remitente:    "test@test.com",
		Destinatario: "dest@test.com",
		Adjunto:      false,
	}, nil)

	emailService := new(MockEmailService)
	emailService.On(
		"SendEmail",
		"test@test.com",
		"dest@test.com",
		"Asunto de prueba",
		"Cuerpo de prueba",
	).Return(nil)

	service := NewPlantillaService(repo, emailService)

	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)

	assert.NoError(t, err, "No debería haber un error cuando la plantilla existe y se envía el correo")
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestHandlePlantilla_ErrorSendingEmail(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Cuerpo de prueba",
		Remitente:    "test@test.com",
		Destinatario: "dest@test.com",
		Adjunto:      false,
	}, nil)

	emailService := new(MockEmailService)
	emailService.On(
		"SendEmail",
		"test@test.com",
		"dest@test.com",
		"Asunto de prueba",
		"Cuerpo de prueba",
	).Return(nil)

	service := NewPlantillaService(repo, emailService)

	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)

	assert.NoError(t, err, "No debería haber un error cuando la plantilla existe y se envía el correo")
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}
