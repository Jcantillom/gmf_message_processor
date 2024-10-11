package logs

import (
	"bytes"
	"os"
	"testing"
)

// Helper para capturar la salida de la consola
func captureOutput(f func()) string {
	var buf bytes.Buffer
	stdout := os.Stdout  // Guardamos el stdout actual
	r, w, _ := os.Pipe() // Creamos una pipe para capturar la salida
	os.Stdout = w        // Redirigimos stdout a la pipe

	f() // Ejecutamos la función que queremos capturar

	_ = w.Close()          // Cerramos la pipe para evitar fugas
	_, _ = buf.ReadFrom(r) // Leemos la salida capturada en el buffer
	os.Stdout = stdout     // Restauramos stdout a su valor original

	return buf.String()
}

func TestLogInfo(t *testing.T) {
	output := captureOutput(func() {
		LogInfo("Este es un mensaje de información", "messageID")
	})

	expected := "INFO"
	if !containsLog(output, expected) {
		t.Errorf("Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s", expected, output)
	}
}

func TestLogWarn(t *testing.T) {
	// Test sin parámetros opcionales
	output := captureOutput(func() {
		LogWarn("Este es un mensaje de advertencia", "messageID")
	})

	expected := "WARNING"
	if !containsLog(output, expected) {
		t.Errorf("Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s", expected, output)
	}

	// Test con parámetros opcionales
	output = captureOutput(func() {
		LogWarn("Este es un mensaje de advertencia", "messageID", "key", "value")
	})

	expected = "WARNING"
	if !containsLog(output, expected) {
		t.Errorf("Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s", expected, output)
	}
}

func TestLogError(t *testing.T) {
	// Test sin error
	output := captureOutput(func() {
		LogError("Este es un error sin detalles", nil, "messageID")
	})

	expected := "ERROR"
	if !containsLog(output, expected) {
		t.Errorf("Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s", expected, output)
	}

	// Test con error
	err := captureOutput(func() {
		LogError("Este es un error con detalles", os.ErrNotExist, "messageID")
	})

	expectedMessage := "Error: file does not exist"
	if !containsLog(err, expectedMessage) {
		t.Errorf("Se esperaba el mensaje '%s' en la salida, pero no se encontró. Salida: %s", expectedMessage, err)
	}
}

func TestLogDebug(t *testing.T) {
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")

	output := captureOutput(func() {
		LogDebug("Este es un mensaje de depuración", "messageID")
	})

	expected := "DEBUG"
	if !containsLog(output, expected) {
		t.Errorf("Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s", expected, output)
	}
}

// Helper para verificar si una cadena contiene el nivel de log esperado
func containsLog(output, expected string) bool {
	return bytes.Contains([]byte(output), []byte(expected))
}
