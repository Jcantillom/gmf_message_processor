package service

import (
	"context"
	"errors"
	"gmf_message_processor/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
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

func (m *MockEmailService) SendEmail(remitente, destinatarios, asunto, cuerpo string) error {
	args := m.Called(remitente, destinatarios, asunto, cuerpo)
	return args.Error(0)
}

func TestHandlePlantilla_InvalidParameters(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(true, &models.Plantilla{
		IDPlantilla:  "PC003",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Cuerpo de prueba",
		Remitente:    "test@example.com",
		Destinatario: "dest@example.com",
		Adjunto:      false,
	}, nil)

	service := NewPlantillaService(repo, nil)

	err := service.HandlePlantilla(context.TODO(), &models.SQSMessage{
		IDPlantilla: "PC003",
		Parametro:   []models.ParametrosSQS{}, // Parámetros vacíos
	})

	assert.Error(t, err, "Debería haber un error cuando no se proporcionan parámetros")
	assert.Equal(t, "no se proporcionaron parámetros para la plantilla", err.Error())
	repo.AssertExpectations(t)
}

func TestHandlePlantilla_PlantillaNotFound(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(false, (*models.Plantilla)(nil), nil)

	service := NewPlantillaService(repo, nil)

	err := service.HandlePlantilla(context.TODO(), &models.SQSMessage{
		IDPlantilla: "PC003",
		Parametro: []models.ParametrosSQS{
			{
				NombreArchivo:      "nombre_archivo",
				DescripcionRechazo: "string",
				CodigoRechazo:      "string",
			},
		},
	})

	assert.Error(t, err, "Debería haber un error cuando la plantilla no existe")
	assert.Equal(t, "la plantilla no existe en la base de datos", err.Error())
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

	err := service.HandlePlantilla(context.TODO(), &models.SQSMessage{
		IDPlantilla: "PC003",
		Parametro: []models.ParametrosSQS{
			{
				NombreArchivo:      "nombre_archivo",
				DescripcionRechazo: "string",
				CodigoRechazo:      "string",
			},
		},
	})

	assert.NoError(t, err, "No debería haber errores")
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)

}

func TestHandlePlantilla_SendEmailError(t *testing.T) {
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
	).Return(errors.New("error enviando el correo electrónico"))

	service := NewPlantillaService(repo, emailService)

	err := service.HandlePlantilla(context.TODO(), &models.SQSMessage{
		IDPlantilla: "PC003",
		Parametro: []models.ParametrosSQS{
			{
				NombreArchivo:      "nombre_archivo",
				DescripcionRechazo: "string",
				CodigoRechazo:      "string",
			},
		},
	})

	assert.Error(t, err, "Debería haber un error cuando el envío de correo falla")
	assert.Equal(t, "error enviando el correo electrónico", err.Error())
	repo.AssertExpectations(t)
	emailService.AssertExpectations(t)
}

func TestHandlePlantilla_RepositoryError(t *testing.T) {
	repo := new(MockPlantillaRepository)
	repo.On("CheckPlantillaExists", "PC003").Return(false, (*models.Plantilla)(nil), errors.New("error de base de datos"))

	service := NewPlantillaService(repo, nil)

	err := service.HandlePlantilla(context.TODO(), &models.SQSMessage{
		IDPlantilla: "PC003",
		Parametro: []models.ParametrosSQS{
			{
				NombreArchivo:      "nombre_archivo",
				DescripcionRechazo: "string",
				CodigoRechazo:      "string",
			},
		},
	})

	assert.Error(t, err, "Debería haber un error cuando el repositorio falla")
	assert.Equal(t, "error de base de datos", err.Error())
	repo.AssertExpectations(t)
}
