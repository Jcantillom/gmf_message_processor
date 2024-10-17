package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// LogInterface define una interfaz para el logger.
type LogInterface interface {
	LogError(message string, err error, messageID string)
	LogInfo(message string, messageID string)
	LogWarn(message string, messageID string, extraArgs ...string)
	LogDebug(message string, messageID string)
}

// LoggerAdapter implementa LogInterface utilizando funciones de logging específicas.
type LoggerAdapter struct{}

func (l *LoggerAdapter) LogError(message string, err error, messageID string) {
	LogError(message, err, messageID)
}

func (l *LoggerAdapter) LogInfo(message, messageID string) {
	LogInfo(message, messageID)
}

func (l *LoggerAdapter) LogWarn(message string, messageID string, extraArgs ...string) {
	LogWarn(message, messageID, extraArgs...)
}

func (l *LoggerAdapter) LogDebug(message string, messageID string) {
	LogDebug(message, messageID)
}

var Logger LogInterface = &LoggerAdapter{}

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

// getCurrentTimestamp obtiene la fecha y hora actual en la zona horaria de Colombia.
func getCurrentTimestamp() string {
	location, err := time.LoadLocation("America/Bogota")
	if err != nil {
		location = time.UTC // Si no se puede cargar, usa UTC como fallback.
	}
	return time.Now().In(location).Format("2006-01-02 15:04:05")
}

// Define una función para obtener el caller, que será usada en los logs
var runtimeCaller = runtime.Caller

// getCallerInfo obtiene el archivo y la línea desde donde se llamó el logger.
func getCallerInfo() string {
	_, file, line, ok := runtimeCaller(3)
	if !ok {
		return "???"
	}
	relativePath := strings.TrimPrefix(filepath.ToSlash(file), basePath)
	return fmt.Sprintf("%s:%d", relativePath, line)
}

// logMessage maneja el formato del log sin colores.
func logMessage(level string, message string, messageID string) {
	timestamp := getCurrentTimestamp()
	callerInfo := getCallerInfo()
	if messageID != "" {
		fmt.Printf(
			"%s [%s] [%s] [MessageId: %s] %s\n",
			timestamp,
			level,
			callerInfo,
			messageID,
			message,
		)
	} else {
		fmt.Printf("%s [%s] [%s] %s\n", timestamp, level, callerInfo, message)
	}
}

// LogInfo genera logs a nivel INFO.
func LogInfo(message string, messageID string) {
	logMessage("INFO", message, messageID)
}

// LogWarn genera logs a nivel WARNING.
func LogWarn(message string, messageID string, extraArgs ...string) {
	if len(extraArgs) >= 2 {
		key := extraArgs[0]
		value := extraArgs[1]
		logMessage("WARNING", fmt.Sprintf("%s - %s: %s", message, key, value), messageID)
	} else {
		logMessage("WARNING", message, messageID)
	}
}

// LogError genera logs a nivel ERROR.
func LogError(message string, err error, messageID string) {
	if err != nil {
		logMessage("ERROR", fmt.Sprintf("%s - Error: %v", message, err), messageID)
	} else {
		logMessage("ERROR", message, messageID)
	}
}

// LogDebug genera logs a nivel DEBUG si el nivel de log es DEBUG.
func LogDebug(message string, messageID string) {
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel == "DEBUG" {
		logMessage("DEBUG", message, messageID)
	}
}
