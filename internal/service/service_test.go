package service

import (
	"context"
	"errors"
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
	).Return(errors.New("error al enviar el correo"))

	service := NewPlantillaService(repo, emailService)

	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)

	assert.Error(t, err)
	assert.Equal(t, "error al enviar el correo", err.Error())

	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)

}

func TestHandlePlantilla_ErrorCheckPlantillaExists(t *testing.T) {
	repo := new(MockPlantillaRepository)
	emailService := new(MockEmailService)

	// Simular un error al verificar si la plantilla existe
	repo.On(
		"CheckPlantillaExists",
		"PC003").Return(false,
		(*models.Plantilla)(nil), errors.New("database error"))

	service := NewPlantillaService(repo, emailService)

	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
		},
		"messageID",
	)

	// Verificar que hubo un error y que el repositorio fue llamado
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	repo.AssertExpectations(t)
}

func TestHandlePlantilla_WithPlaceholders(t *testing.T) {
	repo := new(MockPlantillaRepository)
	emailService := new(MockEmailService)

	// Simular que la plantilla existe en la base de datos
	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Hola, &nombre!",
		Remitente:    "test@example.com",
		Destinatario: "dest@example.com",
		Adjunto:      false,
	}, nil)

	// Simular que el envío de correo es exitoso, incluyendo el `messageID` como quinto argumento
	emailService.On(
		"SendEmail",
		"test@example.com",
		"dest@example.com",
		"Asunto de prueba",
		"Hola, Juan!",
	).Return(nil)

	service := NewPlantillaService(repo, emailService)

	// Llamar a HandlePlantilla con parámetros
	err := service.HandlePlantilla(
		context.TODO(),
		&models.SQSMessage{
			IDPlantilla: "PC003",
			Parametro: []models.ParametrosSQS{
				{Nombre: "nombre", Valor: "Juan"},
			},
		},
		"messageID", // Este es el messageID que debe ser incluido
	)

	// Verificar que no hubo errores y que el correo fue enviado con los placeholders reemplazados
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestHandlePlantilla_PanicOnErrorSendingEmail(t *testing.T) {
	repo := new(MockPlantillaRepository)
	emailService := new(MockEmailService)

	// Simular que la plantilla existe en la base de datos
	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Hola, &nombre!",
		Remitente:    "test@example.com",
		Destinatario: "dest@example.com",
		Adjunto:      false,
	}, nil)

	// Simular que el envío de correo falla
	emailService.On(
		"SendEmail",
		"test@example.com",
		"dest@example.com",
		"Asunto de prueba",
		"Hola, Juan!",
	).Return(errors.New("error al enviar el correo"))

	service := NewPlantillaService(repo, emailService)

	// Llamar a HandlePlantilla con parámetros
	assert.Panics(t, func() {
		_ = service.HandlePlantilla(
			context.TODO(),
			&models.SQSMessage{
				IDPlantilla: "PC003",
				Parametro: []models.ParametrosSQS{
					{Nombre: "nombre", Valor: "Juan"}, // Simular que se pasan parámetros
				},
			},
			"messageID",
		)
	})

	// Verificar que el repositorio y el servicio de correo fueron invocados correctamente
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}
