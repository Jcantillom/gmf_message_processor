package connection

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

// DBManager maneja la conexión y migración de la base de datos.
type DBManager struct {
	DB *gorm.DB
}

// NewDBManager crea una nueva instancia de DBManager.
func NewDBManager() *DBManager {
	return &DBManager{}
}

// InitDB inicializa la conexión a la base de datos y realiza migraciones.
func (dbm *DBManager) InitDB() error {
	// Construir el Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Configurar el logger de GORM
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn, // Cambiado a Warn para ver solo logs importantes
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Abrir la conexión a la base de datos usando GORM
	var err error
	dbm.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // Usar el logger configurado
	})
	if err != nil {
		return fmt.Errorf("error al abrir la conexión a la base de datos: %w", err)
	}

	log.Println("Conexión a la base de datos establecida correctamente 🐘")

	// Migrar la base de datos
	if err := dbm.DB.AutoMigrate(&models.Plantilla{}); err != nil {
		return fmt.Errorf("error al migrar la tabla Plantilla: %w", err)
	}
	log.Println("Migración de la tabla Plantilla completada. 🚀")

	return nil
}

// CloseDB cierra la conexión a la base de datos.
func (dbm *DBManager) CloseDB() {
	sqlDB, err := dbm.DB.DB()
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
