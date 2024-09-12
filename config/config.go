package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
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
		log.Printf("No .env file found, relying on environment variables.")
	}

	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	// Valores predeterminados
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5433")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "gmfdb")
}
