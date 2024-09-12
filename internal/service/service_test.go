package service

import (
	"context"
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
	// Crear instancia del mock del repositorio
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

	// Verificar resultados
	assert.Error(t, err, "Debería haber un error cuando no se proporcionan parámetros")
	assert.Equal(t, "no se proporcionaron parámetros para la plantilla", err.Error())
	repo.AssertExpectations(t)
}

func TestHandlePlantilla_PlantillaNotFound(t *testing.T) {
	// Crear instancia del mock del repositorio
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

	// Verificar resultados
	assert.Error(t, err, "Debería haber un error cuando la plantilla no existe")
	assert.Equal(t, "la plantilla no existe en la base de datos", err.Error())
	repo.AssertExpectations(t)
}
