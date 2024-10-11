package config

import (
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
)

// InitApplication Inicializa la configuración, la conexión a la base de datos y crea las instancias de servicios necesarias.
func InitApplication(
	secretService connection.SecretService, messageID string) (*service.PlantillaService,
	*connection.DBManager, error,
) {
	// Inicializar el ConfigManager y cargar configuración
	configManager := NewConfigManager()
	configManager.InitConfig(messageID)

	// Inicializar el DBManager y abrir la conexión a la base de datos
	dbManager := connection.NewDBManager(secretService)
	err := dbManager.InitDB(messageID)
	if err != nil {
		logs.LogError("Error al inicializar la base de datos", err, messageID)
		return nil, nil, err
	}

	logs.LogDebug("Conexión a la base de datos establecida", messageID)

	// Inicializar el repositorio GORM con la conexión a la base de datos
	repo := repository.NewPlantillaRepository(dbManager.DB)

	// Crear una instancia del servicio de correo electrónico utilizando SMTP
	emailService, smtpErr := email.NewSMTPEmailService(secretService, messageID) // Pasar messageID
	if smtpErr != nil {
		logs.LogError("Error inicializando el servicio SMTP", smtpErr, messageID)
		return nil, nil, smtpErr
	}

	// Crear una instancia del servicio PlantillaService con el servicio de correo electrónico
	plantillaService := service.NewPlantillaService(repo, emailService)

	return plantillaService, dbManager, nil
}

func CleanupApplication(dbManager *connection.DBManager, messageID string) {
	if dbManager != nil {
		// Cerrar la conexión a la base de datos si existe
		dbManager.CloseDB(messageID)
	} else {
		logs.LogError("El DBManager no fue inicializado. No se requiere limpieza.", nil, messageID)
	}
}
