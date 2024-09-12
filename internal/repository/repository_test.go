package repository

import (
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
	repo := NewGormPlantillaRepository(db)

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
