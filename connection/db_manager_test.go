package connection

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDBManager_InitDB(t *testing.T) {
	// Configurar variables de entorno necesarias para la conexión a la base de datos
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "gmfdb")

	// Crear instancia de DBManager
	dbManager := NewDBManager()

	// Inicializar la conexión a la base de datos
	err := dbManager.InitDB()
	assert.NoError(t, err, "La conexión a la base de datos no debería producir un error")

	// Verificar que la conexión no sea nula
	assert.NotNil(t, dbManager.DB, "La instancia DB de Gorm no debería ser nula")

	// Cerrar la conexión de la base de datos
	dbManager.CloseDB()
}
