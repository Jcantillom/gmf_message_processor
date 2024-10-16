package connection

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gmf_message_processor/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ctx = context.TODO()

// DBManagerInterface define los métodos que debe implementar un DBManager.
type DBManagerInterface interface {
	InitDB(messageID string) error
	CloseDB(messageID string)
	GetDB() *gorm.DB
}

// DBManager maneja la conexión y migración de la base de datos.
type DBManager struct {
	DB            *gorm.DB
	SecretService SecretService
}

// NewDBManager crea una nueva instancia de DBManager.
func NewDBManager(service SecretService) *DBManager {
	return &DBManager{
		SecretService: service,
	}
}

// InitDB inicializa la conexión a la base de datos y realiza migraciones.
func (dbm *DBManager) InitDB(messageID string) error {
	// Obtener el secreto
	secretName := os.Getenv("SECRETS_DB")
	secretData, err := dbm.SecretService.GetSecret(secretName, messageID)
	if err != nil {
		logs.LogError("Error al obtener el secreto", err, messageID)
		return fmt.Errorf("error al obtener el secreto: %w", err)
	}

	// Construir el Data Source Name (DSN)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		secretData.Username,
		secretData.Password,
		os.Getenv("DB_NAME"),
	)

	// Configurar el logger de GORM
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Abrir la conexión a la base de datos usando GORM
	if err := dbm.openConnection(postgres.Open(dsn), newLogger, messageID); err != nil {
		return err
	}
	return nil
}

func (dbm *DBManager) openConnection(dialector gorm.Dialector, logger logger.Interface, messageID string) error {
	var err error
	dbm.DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		logs.LogError("Error al abrir la conexión a la base de datos", err, messageID)
		return fmt.Errorf("error al abrir la conexión a la base de datos: %w", err)
	}
	return nil
}

// GetDB obtiene la conexión a la base de datos.
func (dbm *DBManager) GetDB() *gorm.DB {
	return dbm.DB
}

// CloseDB cierra la conexión a la base de datos.
func (dbm *DBManager) CloseDB(messageID string) {
	if dbm.DB == nil {
		// Si la conexión no ha sido inicializada, no intentamos cerrarla
		logs.LogWarn("La conexión a la base de datos no ha sido inicializada", messageID)
		return
	}

	sqlDB, err := dbm.DB.DB()
	if err != nil {
		logs.LogError("Error al obtener la conexión de la base de datos", err, messageID)
		return
	}

	if err := sqlDB.Close(); err != nil {
		logs.LogError("Error al cerrar la conexión de la base de datos", err, messageID)
	} else {
		logs.LogDebug("Conexión a la base de datos cerrada", messageID)
	}
}
