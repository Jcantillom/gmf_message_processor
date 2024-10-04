package logs_test

import (
	"bytes"
	"context"
	"gmf_message_processor/internal/logs"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func captureLogOutput(f func()) string {
	var buf bytes.Buffer
	logs.Log.SetOutput(&buf)
	defer logs.Log.SetOutput(logrus.StandardLogger().Out)

	f()
	return buf.String()
}

func TestLogError(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogError(context.Background(), "Error de prueba: %v", "detalles")
	})
	assert.Contains(t, output, "Error de prueba: detalles")
}

func TestLogInfo(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogInfo(context.Background(), "Información de prueba")
	})
	assert.Contains(t, output, "Información de prueba")
}

func TestLogWarn(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogWarn(context.Background(), "Advertencia de prueba")
	})
	assert.Contains(t, output, "Advertencia de prueba")
}

func TestLogProcesandoMensajeSQS(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogProcesandoMensajeSQS(context.Background())
	})
	assert.Contains(t, output, "Procesando mensaje de SQS... 🚀")
}

func TestLogMensajeProcesadoConExito(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogMensajeProcesadoConExito(context.Background())
	})
	assert.Contains(t, output, "Mensaje procesado con éxito ✅")
}

func TestLogPlantillaNoEncontrada(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogPlantillaNoEncontrada(context.Background(), "PC001")
	})
	assert.Contains(t, output, "La plantilla con ID PC001 no existe en la base de datos  ❌ ")
}

func TestLogPlantillaEncontrada(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogPlantillaEncontrada(context.Background(), "PC001")
	})
	assert.Contains(t, output, "Plantilla con ID PC001 encontrada. Procediendo a enviar el correo electrónico... ✉️")
}

func TestLogCorreoEnviado(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogCorreoEnviado(context.Background(), "PC001")
	})
	assert.Contains(t, output, "Correo electrónico enviado exitosamente para IDPlantilla: PC001 ✅")
}

func TestLogErrorEnvioCorreo(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogErrorEnvioCorreo(context.Background(), "PC001", assert.AnError)
	})
	expected := "Error enviando el correo para la plantilla con ID PC001: assert.AnError general error for testing ❌"
	assert.Contains(t, output, expected)
}

func TestLogParametrosNoProporcionados(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogParametrosNoProporcionados(context.Background(), "PC001")
	})
	assert.Contains(t, output, "No se proporcionaron parámetros para la plantilla con ID PC001")
}

func TestLogFormatoMensajeValido(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogFormatoMensajeValido(context.Background())
	})
	assert.Contains(t, output, "Formato de mensaje válido 😉")
}

func TestLogPlantillaInsertada(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogPlantillaInsertada(context.Background(), "PC001")
	})
	assert.Contains(t, output, "Plantilla con ID PC001 insertada correctamente en la base de datos 🌱")
}

func TestLogDatosSemillaPlantillaInsertados(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogDatosSemillaPlantillaInsertados(context.Background())
	})
	assert.Contains(t, output, "Datos de semilla de plantilla insertados correctamente en la base de datos  🍁")
}

func TestLogEnviandoCorreosADestinatarios(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogEnviandoCorreosADestinatarios(context.Background(), "test@example.com", []byte(`{"message":"hello"}`))
	})
	output = strings.ReplaceAll(output, "\\n", "\n")
	output = strings.ReplaceAll(output, "\\\"", "\"")
	output = strings.TrimSpace(output)
	expected := "Enviando correo electrónico a ... : test@example.com 📤\n{\"message\":\"hello\"}"
	assert.Contains(t, output, expected)
}

func TestLogCorreosEnviados(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogCorreosEnviados(context.Background(), "test@example.com")
	})
	output = strings.ReplaceAll(output, "\\n", "\n")
	output = strings.ReplaceAll(output, "\\\"", "\"")
	output = strings.TrimSpace(output)
	expected := "Correo electrónico enviado con éxito a  ✅ :\ntest@example.com"
	assert.Contains(t, output, expected)
}

func TestLogConexionBaseDatosEstablecida(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogConexionBaseDatosEstablecida()
	})
	assert.Contains(t, output, "Conexión a la base de datos establecida correctamente 🐘")
}

func TestLogErrorConexionBaseDatos(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogErrorConexionBaseDatos(assert.AnError)
	})
	assert.Contains(t, output, "Error al establecer la conexión a la base de datos: assert.AnError")
}

func TestLogErrorMigracionTablaPlantilla(t *testing.T) {
	output := captureLogOutput(func() {
		err := logs.LogErrorMigracionTablaPlantilla(assert.AnError)
		assert.Equal(t, assert.AnError, err)
	})
	assert.Contains(t, output, "Error al migrar la tabla Plantilla: assert.AnError")
}

func TestLogErrorCerrandoConexionBaseDatos(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogErrorCerrandoConexionBaseDatos(assert.AnError)
	})
	assert.Contains(t, output, "Error al cerrar la conexión a la base de datos: assert.AnError")
}

func TestLogConexionBaseDatosCerrada(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogConexionBaseDatosCerrada()
	})
	assert.Contains(t, output, "Conexión a la base de datos cerrada correctamente 🚪")
}

func TestLogMigracionTablaPlantillaCompletada(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogMigracionTablaPlantillaCompletada()
	})
	assert.Contains(t, output, "Migración de la tabla Plantilla completada. 🚀")
}

func TestLogArchivoEnvNoEncontrado(t *testing.T) {
	output := captureLogOutput(func() {
		logs.LogArchivoEnvNoEncontrado()
	})
	assert.Contains(t, output, "No se encontró el archivo .env, confiando en las variables de entorno.")
}
