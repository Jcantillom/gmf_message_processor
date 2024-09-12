package handler

import (
	"context"
	"github.com/stretchr/testify/mock"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/service"
	"gmf_message_processor/internal/utils"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

type ProcessSQSMessagesTest struct {
	name             string
	ctx              context.Context
	plantillaService *service.PlantillaService
	wantProcessed    int
	wantErr          error
}

type MockPlantillaRepository struct {
	mock.Mock
}

func (m *MockPlantillaRepository) CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error) {
	args := m.Called(idPlantilla)
	return args.Bool(0), args.Get(1).(*models.Plantilla), args.Error(2)
}

func (test *ProcessSQSMessagesTest) Run(t *testing.T) {
	t.Helper()

	// Obtener el directorio de trabajo actual
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error obteniendo el directorio de trabajo: %v", err)
	}

	// Construir la ruta absoluta al archivo .env
	envPath := filepath.Join(workingDir, "../../.env")

	// Cargar archivo .env
	err = godotenv.Load(envPath)
	if err != nil {
		t.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	// Leer el evento SQS desde el archivo JSON usando la función ReadSQSEventFromFile de utils.
	sqsEvent, err := utils.ReadSQSEventFromFile("../../test_data/message.json")
	if err != nil {
		t.Errorf("Error al leer el evento SQS: %v", err)
		return
	}

	processed := 0 // Contador de mensajes procesados

	// Procesar cada mensaje en el evento
	for _, message := range sqsEvent.Records {
		// Validar el mensaje
		validMsg, err := utils.ValidateSQSMessage(message.Body)
		if err != nil {
			t.Errorf("Error validando el mensaje ❌: %v", err)
			continue
		}

		// Llamar al servicio para manejar la lógica de negocio
		if err := test.plantillaService.HandlePlantilla(test.ctx, validMsg); err != nil {
			t.Errorf("Error al procesar el mensaje para IDPlantilla: %s: %v", validMsg.IDPlantilla, err)
			return
		} else {
			t.Logf("Mensaje procesado con éxito para IDPlantilla ✅: %s", validMsg.IDPlantilla)
			processed++ // Incrementar el contador de mensajes procesados
		}
	}

	if processed != test.wantProcessed {
		t.Errorf("Procesados = %d, want %d", processed, test.wantProcessed)
	}
}

func TestProcessSQSMessages(t *testing.T) {
	// Crear instancia del mock del repositorio
	repo := new(MockPlantillaRepository)

	// Mockear la respuesta del repositorio
	repo.On("CheckPlantillaExists", "PC001").Return(true, &models.Plantilla{
		IDPlantilla:  "PC001",
		Asunto:       "Asunto de prueba",
		Cuerpo:       "Cuerpo con &nombre_archivo",
		Remitente:    "test@example.com",
		Destinatario: "dest@example.com",
		Adjunto:      false,
	}, nil)

	// Crear instancia del servicio con el mock
	plantillaService := service.NewPlantillaService(repo, nil)

	tests := []ProcessSQSMessagesTest{
		{
			name:             "Procesar mensajes SQS",
			ctx:              context.Background(),
			plantillaService: plantillaService, // Usar el servicio inicializado
			wantProcessed:    1,
			wantErr:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.Run)
	}
}
