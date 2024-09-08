package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"gmf_message_processor/connections"
	"log"
)

// InitConfig initializes the configuration by loading environment variables.
func InitConfig() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables.")
	}

	viper.SetEnvPrefix("app")
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5433")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "gmfdb")

	// Initialize the database connection
	connections.InitDB()
	// Insertar datos de semilla en la base de datos.
	//seeds.SeedDataPlantilla()

}
