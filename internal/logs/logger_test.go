package logs

import (
	"bytes"
	"os"
	"testing"
)

// Helper para verificar si una cadena contiene el nivel de log esperado
func containsLog(output, expected string) bool {
	return bytes.Contains([]byte(output), []byte(expected))
}

const (
	mensajeInfo  = "Este es un mensaje de información"
	mensajeError = "Se esperaba el nivel %s en la salida, pero no se encontró. Salida: %s"
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
		LogInfo(mensajeInfo, "messageID")
	})

	expected := "INFO"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLogWarn(t *testing.T) {
	// Test sin parámetros opcionales
	output := captureOutput(func() {
		LogWarn("Este es un mensaje de advertencia", "messageID")
	})

	expected := "WARNING"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}

	// Test con parámetros opcionales
	output = captureOutput(func() {
		LogWarn("Este es un mensaje de advertencia", "messageID", "key", "value")
	})

	expected = "WARNING"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLogError(t *testing.T) {
	// Test sin error
	output := captureOutput(func() {
		LogError("Este es un error sin detalles", nil, "messageID")
	})

	expected := "ERROR"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}

	// Test con error
	err := captureOutput(func() {
		LogError("Este es un error con detalles", os.ErrNotExist, "messageID")
	})

	expectedMessage := "Error: file does not exist"
	if !containsLog(err, expectedMessage) {
		t.Errorf(mensajeError, expectedMessage, err)
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
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLoggerAdapterLogError(t *testing.T) {
	adapter := &LoggerAdapter{}
	output := captureOutput(func() {
		adapter.LogError("Este es un error desde el adaptador", os.ErrNotExist, "messageID")
	})

	expected := "ERROR"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}

	expectedMessage := "Error: file does not exist"
	if !containsLog(output, expectedMessage) {
		t.Errorf(mensajeError, expectedMessage, output)
	}
}

func TestLoggerAdapterLogInfo(t *testing.T) {
	adapter := &LoggerAdapter{}
	output := captureOutput(func() {
		adapter.LogInfo("Este es un mensaje de información desde el adaptador", "messageID")
	})

	expected := "INFO"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLoggerAdapterLogWarn(t *testing.T) {
	adapter := &LoggerAdapter{}
	output := captureOutput(func() {
		adapter.LogWarn("Este es un mensaje de advertencia desde el adaptador", "messageID")
	})

	expected := "WARNING"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLoggerAdapterLogDebug(t *testing.T) {
	adapter := &LoggerAdapter{}
	os.Setenv("LOG_LEVEL", "DEBUG")
	defer os.Unsetenv("LOG_LEVEL")

	output := captureOutput(func() {
		adapter.LogDebug("Este es un mensaje de depuración desde el adaptador", "messageID")
	})

	expected := "DEBUG"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
}

func TestLogMessageWithoutMessageID(t *testing.T) {
	output := captureOutput(func() {
		logMessage("INFO", mensajeInfo, "")
	})

	expected := "INFO"
	if !containsLog(output, expected) {
		t.Errorf(mensajeError, expected, output)
	}
	expectedMessage := "Este es un mensaje de información"
	if !containsLog(output, expectedMessage) {
		t.Errorf(mensajeError, expectedMessage, output)
	}
}

func TestLogMessageWithMessageID(t *testing.T) {
	output := captureOutput(func() {
		logMessage("INFO", mensajeInfo, "messageID")
	})

	expected := "[MessageId: messageID]"
	if !containsLog(output, expected) {
		t.Errorf(
			"Se esperaba el MessageId 'messageID' en la salida, pero no se encontró. Salida: %s",
			output,
		)
	}
}

func TestGetCallerInfoFailure(t *testing.T) {
	// Simula el fallo de runtime.Caller
	originalCaller := runtimeCaller
	runtimeCaller = func(skip int) (pc uintptr, file string, line int, ok bool) {
		return 0, "", 0, false
	}
	defer func() { runtimeCaller = originalCaller }() // Restaurar la función original.

	callerInfo := getCallerInfo()
	if callerInfo != "???" {
		t.Errorf("Se esperaba que callerInfo fuera '???', pero fue: %s", callerInfo)
	}
}
