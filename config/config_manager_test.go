package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfigManager_InitConfig(t *testing.T) {
	// Crear instancia de ConfigManager
	configManager := NewConfigManager()

	// Llamar a InitConfig para inicializar la configuración
	configManager.InitConfig()

	// Verificar que las variables de entorno se han cargado correctamente
	assert.Equal(t, "localhost", viper.GetString("DB_HOST"))
	assert.Equal(t, "5433", viper.GetString("DB_PORT"))
	assert.Equal(t, "postgres", viper.GetString("DB_USER"))
	assert.Equal(t, "postgres", viper.GetString("DB_PASSWORD"))
	assert.Equal(t, "gmfdb", viper.GetString("DB_NAME"))
}
