package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Definir colores ANSI
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
)

var basePath string

// Inicializa la ruta base al inicio
func init() {
	// Obtiene el directorio actual de trabajo
	wd, err := os.Getwd()
	if err != nil {
		basePath = ""
	} else {
		// Normaliza la ruta
		basePath = filepath.ToSlash(wd)
	}
}

// Función para obtener la fecha actual en formato deseado
func getCurrentTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// Función para obtener el archivo y la línea de donde se llamó el log
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "???"
	}
	// Eliminar la ruta base del sistema para mostrar solo la parte relativa del proyecto
	relativePath := strings.TrimPrefix(filepath.ToSlash(file), basePath)
	return fmt.Sprintf("%s:%d", relativePath, line)
}

// Función genérica para loguear un mensaje donde solo el nivel de log está coloreado
func logMessage(levelColor string, level string, message string) {
	timestamp := getCurrentTimestamp()
	callerInfo := getCallerInfo()
	// Ajustamos el orden para que el nivel esté después de la fecha y antes de la ruta
	fmt.Printf("%s [%s%s%s] [%s] %s\n", timestamp, levelColor, level, Reset, callerInfo, message)
}

// LogInfo genera un log a nivel INFO con nivel coloreado en verde
func LogInfo(message string) {
	logMessage(Green, "INFO", message)
}

// LogWarn genera un log a nivel WARNING con nivel coloreado en amarillo
func LogWarn(message string, extraArgs ...string) {
	// Verificar si se pasaron el key y value opcionalmente
	if len(extraArgs) >= 2 {
		key := extraArgs[0]
		value := extraArgs[1]
		logMessage(Yellow, "WARNING", fmt.Sprintf("%s - %s: %s", message, key, value))
	} else {
		// Solo muestra el mensaje sin parámetros adicionales
		logMessage(Yellow, "WARNING", message)
	}
}

// LogError genera un log a nivel ERROR con nivel coloreado en rojo
func LogError(message string, err error) {
	if err != nil {
		// Incluir el detalle del error en el log si está disponible
		logMessage(Red, "ERROR", fmt.Sprintf("%s - Error: %v", message, err))
	} else {
		logMessage(Red, "ERROR", message)
	}
}

// LogDebug genera un log a nivel DEBUG con nivel coloreado en azul
func LogDebug(message string) {
	logMessage(Blue, "DEBUG", message)
}
