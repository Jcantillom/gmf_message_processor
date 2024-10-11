package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock de godotenv para simular errores
type MockGodotenv struct {
	mock.Mock
}

func (m *MockGodotenv) Load() error {
	args := m.Called()
	return args.Error(0)
}

func TestConfigManager_UnsetConfig(t *testing.T) {
	// Limpiar cualquier configuración previa
	viper.Reset()

	// Verificar que las configuraciones no establecidas devuelvan el valor predeterminado
	assert.Equal(t, "", viper.GetString("UNSET_CONFIG"))
}

func TestConfigManager_UpdateConfig(t *testing.T) {
	// Establecer valores simulados de configuración usando viper.Set()
	viper.Set("DB_HOST", "localhost")

	// Verificar el valor inicial
	assert.Equal(t, "localhost", viper.GetString("DB_HOST"))

	// Actualizar la configuración
	viper.Set("DB_HOST", "newhost")

	// Verificar el valor actualizado
	assert.Equal(t, "newhost", viper.GetString("DB_HOST"))
}

func TestConfigManager_ResetConfig(t *testing.T) {
	// Establecer valores simulados de configuración usando viper.Set()
	viper.Set("DB_HOST", "localhost")

	// Verificar que la configuración esté establecida
	assert.Equal(t, "localhost", viper.GetString("DB_HOST"))

	// Resetear la configuración
	viper.Reset()

	// Verificar que la configuración se haya limpiado
	assert.Equal(t, "", viper.GetString("DB_HOST"))
}
