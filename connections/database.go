package connections

import (
	"fmt"
	"gmf_message_processor/internal/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB inicializa la conexión a la base de datos y configura GORM.
func InitDB() {
	// Construir el Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Configurar el logger de GORM a nivel Silent
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Abrir la conexión a la base de datos usando GORM
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // Usar el logger configurado
	})
	if err != nil {
		log.Fatalf("Error al abrir la conexión a la base de datos: %v", err)
	}

	log.Println("Conexión a la base de datos establecida correctamente 🐘")

	// Migrar la base de datos
	if err := DB.AutoMigrate(&models.Plantilla{}); err != nil {
		log.Fatalf("Error al migrar la tabla Plantilla: %v", err)
	}
	log.Println("Migración de la tabla Plantilla completada. 🚀")
}

// CloseDB cierra la conexión a la base de datos.
func CloseDB() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Error obteniendo la instancia SQL DB de GORM: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Error al cerrar la conexión a la base de datos: %v", err)
	} else {
		log.Println("Conexión a la base de datos cerrada.")
	}
}
