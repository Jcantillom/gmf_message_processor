package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gmf_message_processor/internal/logs"
	"log"
)

// Manager ConfigManager maneja la carga de configuración de la aplicación.
type Manager struct{}

// NewConfigManager crea una nueva instancia de ConfigManager.
func NewConfigManager() *Manager {
	return &Manager{}
}

// InitConfig inicializa la configuración de la aplicación.
func (cm *Manager) InitConfig(messageID string) {
	if err := godotenv.Load(); err != nil {
		logs.LogDebug(
			"Archivo .env no encontrado, se utilizarán las variables de entorno del sistema",
			messageID,
		)
	} else {
		logs.LogDebug(
			"Leyendo variables de entorno desde el archivo .env",
			messageID,
		)
	}

	// Configurar Viper para cargar automáticamente las variables de entorno
	viper.AutomaticEnv()

	// Si existe el archivo .env, Viper lo cargará. Si no, continuará con las variables de entorno del sistema
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		logs.LogWarn(
			"No se pudo cargar el archivo .env, se utilizarán las variables de entorno del sistema",
			messageID,
		)
	}

	// Definir las variables clave que deben estar presentes en el entorno
	requiredEnvVars := []string{
		"APP_ENV",
		"SERVICE_ENV",
		"SECRETS_DB",
		"SECRETS_SMTP",
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"DB_SCHEMA",
		"SMTP_SERVER",
		"SMTP_PORT",
		"SQS_QUEUE_URL",
		"SMTP_TIMEOUT",
	}

	// Verificar si las variables clave están presentes
	for _, key := range requiredEnvVars {
		if !viper.IsSet(key) {
			logs.LogError("La variable de entorno "+key+" no está configurada", nil, messageID)
			log.Fatalf("**** Revise la configuración de la aplicación ****")
		}
	}
}

func setDefault(key, value string, messageID string) {
	if !viper.IsSet(key) {
		viper.SetDefault(key, value)
		logs.LogInfo("Estableciendo valor por defecto para "+key, messageID)
	}
}
