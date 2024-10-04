package repository

import (
	"errors"
	"gmf_message_processor/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckPlantillaExists(t *testing.T) {
	// Crear una base de datos en memoria usando SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error al conectar a la base de datos en memoria: %v", err)
	}

	// Limpiar después de la prueba
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("Error al obtener instancia de DB para cerrar la conexión: %v", err)
		}
		sqlDB.Close()
	})

	// Crear tabla de prueba
	err = db.AutoMigrate(&models.Plantilla{})
	if err != nil {
		t.Fatalf("Error al migrar tabla de prueba: %v", err)
	}

	// Crear instancia de GormPlantillaRepository
	repo := NewPlantillaRepository(db)

	// Insertar plantilla de prueba
	plantilla := models.Plantilla{
		IDPlantilla: "plantilla-1",
		Asunto:      "Asunto de Prueba",
		Cuerpo:      "Cuerpo de Prueba",
	}
	err = db.Create(&plantilla).Error
	if err != nil {
		t.Fatalf("Error al insertar plantilla de prueba: %v", err)
	}

	// Verificar que la plantilla exista
	existe, _, err := repo.CheckPlantillaExists("plantilla-1")
	if err != nil {
		t.Fatalf("Error al verificar la existencia de la plantilla: %v", err)
	}
	if !existe {
		t.Fatalf("La plantilla debería existir en la base de datos")
	}

	// Verificar que la plantilla no exista
	existe, _, err = repo.CheckPlantillaExists("plantilla-2")
	if err != nil {
		t.Fatalf("Error al verificar la existencia de la plantilla: %v", err)
	}

	if existe {
		t.Fatalf("La plantilla no debería existir en la base de datos")
	}
}

func TestCheckPlantillaExists_Error(t *testing.T) {
	// Crear una base de datos en memoria usando SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error al conectar a la base de datos en memoria: %v", err)
	}

	// Crear instancia de GormPlantillaRepository
	repo := NewPlantillaRepository(db)

	// Simular un error en la consulta utilizando el método de callback de GORM
	db.Callback().Query().Replace("gorm:query", func(tx *gorm.DB) {
		tx.AddError(errors.New("error simulado"))
	})

	// Verificar que se produce un error inesperado
	_, _, err = repo.CheckPlantillaExists("plantilla-error")
	if err == nil || err.Error() != "error simulado" {
		t.Fatalf("Se esperaba un error simulado, pero se obtuvo: %v", err)
	}
}

func TestCheckPlantillaExists_DatabaseConnectionError(t *testing.T) {
	// Crear una base de datos en memoria usando SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error al conectar a la base de datos en memoria: %v", err)
	}

	// Limpiar después de la prueba
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("Error al obtener instancia de DB para cerrar la conexión: %v", err)
		}
		sqlDB.Close()
	})

	// Cerrar la conexión a la base de datos para simular el error
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Error al obtener instancia de DB para cerrar la conexión: %v", err)
	}
	sqlDB.Close()

	// Crear instancia de GormPlantillaRepository con la conexión cerrada
	repo := NewPlantillaRepository(db)

	// Verificar que se produce un error al intentar verificar la existencia de una plantilla
	_, _, err = repo.CheckPlantillaExists("plantilla-1")
	if err == nil {
		t.Fatalf("Se esperaba un error de conexión a la base de datos, pero no se produjo ninguno")
	}
}
