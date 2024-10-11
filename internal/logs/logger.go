package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogInterface define una interfaz para el logger, lo que permite fácilmente intercambiar implementaciones.
type LogInterface interface {
	LogError(message string, err error, messageID string)
	LogInfo(message string, messageID string)
	LogWarn(message string, messageID string, extraArgs ...string)
	LogDebug(message string, messageID string)
}

// LoggerAdapter implementa LogInterface utilizando funciones de logging específicas.
type LoggerAdapter struct{}

// LogError adapta la función LogError para cumplir con la interfaz LogInterface.
func (l *LoggerAdapter) LogError(message string, err error, messageID string) {
	LogError(message, err, messageID)
}

// LogInfo adapta la función LogInfo para cumplir con la interfaz LogInterface.
func (l *LoggerAdapter) LogInfo(message, messageID string) {
	LogInfo(message, messageID)
}

// LogWarn adapta la función LogWarn para cumplir con la interfaz LogInterface.
func (l *LoggerAdapter) LogWarn(message string, messageID string, extraArgs ...string) {
	LogWarn(message, messageID, extraArgs...)
}

// LogDebug adapta la función LogDebug para cumplir con la interfaz LogInterface.
func (l *LoggerAdapter) LogDebug(message string, messageID string) {
	LogDebug(message, messageID)
}

// Definir colores ANSI para el terminal
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

// basePath almacena la ruta base del proyecto, que se obtiene al iniciar la aplicación.
var basePath string

// Inicializa la ruta base al inicio
func init() {
	wd, err := os.Getwd()
	if err != nil {
		basePath = ""
	} else {
		basePath = filepath.ToSlash(wd)
	}
}

// getCurrentTimestamp obtiene la fecha y hora actual en un formato legible.
func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// getCallerInfo obtiene el archivo y la línea desde donde se llamó el logger.
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "???"
	}
	relativePath := strings.TrimPrefix(filepath.ToSlash(file), basePath)
	return fmt.Sprintf("%s:%d", relativePath, line)
}

// logMessage es una función genérica que maneja el formato del log y el mensaje.
func logMessage(levelColor string, level string, message string, messageID string) {
	timestamp := getCurrentTimestamp()
	callerInfo := getCallerInfo()
	if messageID != "" {
		fmt.Printf("%s [%s%s%s] [%s] [MessageId: %s] %s\n", timestamp, levelColor, level, Reset, callerInfo, messageID, message)
	} else {
		fmt.Printf("%s [%s%s%s] [%s] %s\n", timestamp, levelColor, level, Reset, callerInfo, message)
	}
}

// LogInfo genera logs a nivel INFO (en verde).
func LogInfo(message string, messageID string) {
	logMessage(Green, "INFO", message, messageID)
}

// LogWarn genera logs a nivel WARNING (en amarillo).
func LogWarn(message string, messageID string, extraArgs ...string) {
	if len(extraArgs) >= 2 {
		key := extraArgs[0]
		value := extraArgs[1]
		logMessage(Yellow, "WARNING", fmt.Sprintf("%s - %s: %s", message, key, value), messageID)
	} else {
		logMessage(Yellow, "WARNING", message, messageID)
	}
}

// LogError genera logs a nivel ERROR (en rojo).
func LogError(message string, err error, messageID string) {
	if err != nil {
		logMessage(Red, "ERROR", fmt.Sprintf("%s - Error: %v", message, err), messageID)
	} else {
		logMessage(Red, "ERROR", message, messageID)
	}
}

// LogDebug genera logs a nivel DEBUG (en azul) solo si el nivel de log es DEBUG.
func LogDebug(message string, messageID string) {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel == "DEBUG" {
		logMessage(Blue, "DEBUG", message, messageID)
	}
}
