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

// InitConfig carga la configuración desde el archivo .env si existe, o desde las variables de entorno del sistema.
func (cm *Manager) InitConfig() {
	// Intentar cargar el archivo .env, pero si no existe, continuar con las variables de entorno
	err := godotenv.Load()
	if err != nil {
		logs.LogWarn("usando variables de entorno")
	}
	logs.LogWarn("usando variables de entorno del archivo .env")
	// Configurar Viper para cargar automáticamente las variables de entorno
	viper.AutomaticEnv()

	// Si el archivo .env existe, intentará cargarlo, pero si no, Viper seguirá con las variables de entorno
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		logs.LogInfo("No se pudo cargar el archivo .env")
	} else {
		logs.LogInfo("Leyendo Variables de entorno desde el archivo .env")
	}

	// Definir las variables clave que deben estar presentes en el entorno
	requiredEnvVars := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_NAME",
		"SMTP_SERVER",
		"SMTP_PORT",
		"SQS_QUEUE_URL",
		"EMAIL_PROVIDER",
	}

	// Verificar si las variables clave están presentes
	for _, key := range requiredEnvVars {
		if !viper.IsSet(key) {
			logs.LogError("La variable de entorno %s no está configurada", err)
			log.Fatalf("**** Revise la configuración de la aplicación ****")

		}
	}

	// Enlazar variables de entorno con Viper (para asegurarnos de que están disponibles en el entorno de Lambda o local)
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("SMTP_SERVER")
	viper.BindEnv("SMTP_PORT")
	viper.BindEnv("SMTP_USER")
	viper.BindEnv("SMTP_PASSWORD")
	viper.BindEnv("SQS_QUEUE_URL")
	viper.BindEnv("SECRET_NAME")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("EMAIL_PROVIDER")
	viper.BindEnv("SES_SENDER_EMAIL")
}

func setDefault(key, value string) {
	if !viper.IsSet(key) {
		viper.SetDefault(key, value)
	}
}
