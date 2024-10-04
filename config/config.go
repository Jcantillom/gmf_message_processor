package config

import (
	"gmf_message_processor/internal/logs"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Manager ConfigManager maneja la carga de configuración de la aplicación.
type Manager struct{}

// NewConfigManager crea una nueva instancia de ConfigManager.
func NewConfigManager() *Manager {
	return &Manager{}
}

// InitConfig carga la configuración desde el archivo .env y las variables de entorno.
func (cm *Manager) InitConfig() {
	// Cargar el archivo .env si existe
	if err := godotenv.Load(); err != nil {
		logs.LogArchivoEnvNoEncontrado()
	}

	// Configurar Viper para usar el prefijo "APP" y cargar variables automáticamente del entorno
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	// Especificar el archivo .env y cargarlo con Viper
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Leer las configuraciones del archivo .env
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error al leer el archivo de configuración: %v", err)
	}

	// Enlazar variables de entorno con Viper
	viper.BindEnv("DB_HOST")
	viper.BindEnv("DB_PORT")
	viper.BindEnv("DB_USER")
	viper.BindEnv("DB_PASSWORD")
	viper.BindEnv("DB_NAME")
	viper.BindEnv("SMTP_SERVER")
	viper.BindEnv("SMTP_PORT")
	viper.BindEnv("SMTP_USER")
	viper.BindEnv("SMTP_PASSWORD")
	viper.BindEnv("SQS_QUEUE_URL")

}

func setDefault(key, value string) {
	if !viper.IsSet(key) {
		viper.SetDefault(key, value)
	}
}
