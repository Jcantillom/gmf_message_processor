package connection

import (
	"context"
	"fmt"
	"gmf_message_processor/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
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
	Logger        logger.Interface
}

// NewDBManager crea una nueva instancia de DBManager.
func NewDBManager(service SecretService, logger logger.Interface) *DBManager {
	return &DBManager{
		SecretService: service,
		Logger:        logger,
	}
}

// InitDB inicializa la conexión a la base de datos y realiza migraciones.
func (dbm *DBManager) InitDB(messageID string) error {
	secretName := os.Getenv("SECRETS_DB")
	secretData, err := dbm.SecretService.GetSecret(secretName, messageID)
	if err != nil {
		logs.LogError("Error al obtener el secreto", err, messageID)
		return fmt.Errorf("error al obtener el secreto: %w", err)
	}

	dsn := buildDSN(secretData)
	if err := dbm.openConnection(postgres.Open(dsn), dbm.Logger, messageID); err != nil {
		return fmt.Errorf("error al abrir la conexión: %w", err)
	}
	return nil
}

// buildDSN construye el DSN de la base de datos.
func buildDSN(secretData *SecretData) string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		secretData.Username,
		secretData.Password,
		os.Getenv("DB_NAME"),
	)
}

// openConnection abre la conexión con la base de datos.
func (dbm *DBManager) openConnection(dialector gorm.Dialector, logger logger.Interface, messageID string) error {
	var err error
	dbm.DB, err = gorm.Open(dialector, &gorm.Config{Logger: logger})
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
