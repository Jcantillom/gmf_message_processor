package connection

import (
	"fmt"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBManagerInterface define los métodos que debe implementar un DBManager.
type DBManagerInterface interface {
	InitDB() error
	CloseDB()
	GetDB() *gorm.DB
}

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
	// Obtener el secreto
	secretName := os.Getenv("SECRET_NAME") // Nombre del secreto en AWS
	secretData, err := getSecret(secretName)
	if err != nil {
		return fmt.Errorf("error al obtener el secreto: %w", err)
	}

	// Construir el Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		secretData.Username,
		secretData.Password,
		os.Getenv("DB_NAME"),
	)

	// Configurar el logs de GORM
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
	dbm.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger, // Usar el logs configurado
	})
	if err != nil {
		return fmt.Errorf("error al abrir la conexión a la base de datos: %w", err)
	}

	logs.LogConexionBaseDatosEstablecida()

	// Migrar la base de datos
	if err := dbm.DB.AutoMigrate(&models.Plantilla{}); err != nil {
		return logs.LogErrorMigracionTablaPlantilla(err)
	}
	logs.LogMigracionTablaPlantillaCompletada()

	return nil
}

// GetDB obtiene la conexión a la base de datos.
func (dbm *DBManager) GetDB() *gorm.DB {
	return dbm.DB
}

// CloseDB cierra la conexión a la base de datos.
func (dbm *DBManager) CloseDB() {
	sqlDB, err := dbm.DB.DB()
	if err != nil {
		logs.LogErrorConexionBaseDatos(err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		logs.LogErrorCerrandoConexionBaseDatos(err)
	} else {
		logs.LogConexionBaseDatosCerrada()
	}
}
