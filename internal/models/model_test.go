package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPlantillaModel_Initialization(t *testing.T) {
	// Crear una instancia de Plantilla con datos de prueba
	plantilla := Plantilla{
		IDPlantilla:  "12345678-1234-1234-1234-123456789012",
		Asunto:       "Prueba de Asunto",
		Cuerpo:       "Este es el cuerpo del correo de prueba.",
		Remitente:    "remitente@ejemplo.com",
		Destinatario: "destinatario@ejemplo.com",
		Adjunto:      false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Verificar que los campos de Plantilla se han inicializado correctamente
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", plantilla.IDPlantilla)
	assert.Equal(t, "Prueba de Asunto", plantilla.Asunto)
	assert.Equal(t, "Este es el cuerpo del correo de prueba.", plantilla.Cuerpo)
	assert.Equal(t, "remitente@ejemplo.com", plantilla.Remitente)
	assert.Equal(t, "destinatario@ejemplo.com", plantilla.Destinatario)
	assert.Equal(t, false, plantilla.Adjunto)
	assert.NotZero(t, plantilla.CreatedAt)
	assert.NotZero(t, plantilla.UpdatedAt)
}

func TestPlantillaModel_TableName(t *testing.T) {
	// Verificar que el nombre de la tabla sea el esperado
	var plantilla Plantilla
	assert.Equal(t, "CGD_CORREOS_PLANTILLAS", plantilla.TableName())
}

func TestSQSMessageModel_Initialization(t *testing.T) {
	// Crear una instancia de SQSMessage con datos de prueba
	sqsMessage := SQSMessage{
		IDPlantilla: "12345678-1234-1234-1234-123456789012",
		Parametro: []ParametrosSQS{
			{
				NombreArchivo:      "archivo1.txt",
				PlataformaOrigen:   "Sistema X",
				FechaRecepcion:     "2024-09-12",
				HoraRecepcion:      "10:00",
				CodigoRechazo:      "404",
				DescripcionRechazo: "Archivo no encontrado",
				DetalleRechazo:     "El archivo solicitado no está disponible en el servidor.",
			},
		},
	}

	// Verificar que los campos de SQSMessage se han inicializado correctamente
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", sqsMessage.IDPlantilla)
	assert.Len(t, sqsMessage.Parametro, 1)

	// Verificar que los campos de ParametrosSQS se han inicializado correctamente
	param := sqsMessage.Parametro[0]
	assert.Equal(t, "archivo1.txt", param.NombreArchivo)
	assert.Equal(t, "Sistema X", param.PlataformaOrigen)
	assert.Equal(t, "2024-09-12", param.FechaRecepcion)
	assert.Equal(t, "10:00", param.HoraRecepcion)
	assert.Equal(t, "404", param.CodigoRechazo)
	assert.Equal(t, "Archivo no encontrado", param.DescripcionRechazo)
	assert.Equal(t, "El archivo solicitado no está disponible en el servidor.", param.DetalleRechazo)
}
