package models

import (
	"os"
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
	assert.Equal(t, "cgd_correos_plantillas", plantilla.TableName())
}

func TestSQSMessageModel_Initialization(t *testing.T) {
	// Crear una instancia de SQSMessage con datos de prueba
	sqsMessage := SQSMessage{
		IDPlantilla: "12345678-1234-1234-1234-123456789012",
		Parametro: []ParametrosSQS{
			{
				Nombre: "archivo1.txt",
				Valor:  "Sistema X",
			},
		},
	}

	// Verificar que los campos de SQSMessage se han inicializado correctamente
	assert.Equal(t, "12345678-1234-1234-1234-123456789012", sqsMessage.IDPlantilla)
	assert.Len(t, sqsMessage.Parametro, 1)

	// Verificar que los campos de ParametrosSQS se han inicializado correctamente
	param := sqsMessage.Parametro[0]
	assert.Equal(t, "archivo1.txt", param.Nombre)
	assert.Equal(t, "Sistema X", param.Valor)

}
func TestPlantillaTableName_DefaultSchema(t *testing.T) {
	// Limpiar la variable de entorno para simular el comportamiento por defecto
	os.Unsetenv("DB_SCHEMA")

	plantilla := Plantilla{}
	expectedTableName := "cgd_correos_plantillas"

	tableName := plantilla.TableName()

	assert.Equal(
		t,
		expectedTableName,
		tableName,
		"El nombre de la tabla por defecto debería ser 'cgd_correos_plantillas'")
}

func TestPlantillaTableName_CustomSchema(t *testing.T) {
	// Configurar una variable de entorno para simular un esquema personalizado
	os.Setenv("DB_SCHEMA", "custom_schema")

	plantilla := Plantilla{}
	expectedTableName := "custom_schema.cgd_correos_plantillas"

	tableName := plantilla.TableName()

	assert.Equal(
		t,
		expectedTableName,
		tableName,
		"El nombre de la tabla con esquema personalizado debería ser 'custom_schema.cgd_correos_plantillas'")

	// Limpiar la variable de entorno después del test
	os.Unsetenv("DB_SCHEMA")
}
