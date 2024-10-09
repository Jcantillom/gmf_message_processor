package config

import (
	"context"
	"github.com/spf13/viper"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
	"time"
)

var ctx = context.TODO()

// Reintentar la inicialización en caso de fallo (por ejemplo, problemas de conectividad temporal)
func retry(fn func() error, retries int, delay time.Duration) error {
	var err error
	for i := 0; i < retries; i++ {
		if err = fn(); err == nil {
			return nil
		}
		logs.LogWarn("Reintento %d fallido: %v. Intentando de nuevo...")
		time.Sleep(delay)
	}
	return err
}

// Inicializa la configuración, la conexión a la base de datos y crea las instancias de servicios necesarias.
func InitApplication(
	secretService connection.SecretService) (*service.PlantillaService,
	*connection.DBManager, error,
) {
	// Inicializar el ConfigManager y cargar configuración
	configManager := NewConfigManager()
	configManager.InitConfig()

	// Inicializar el DBManager y abrir la conexión a la base de datos
	dbManager := connection.NewDBManager(secretService)
	err := retry(dbManager.InitDB, 3, 5*time.Second)
	if err != nil {
		return nil, nil, err
	}

	logs.LogInfo("Conexión a la base de datos establecida")

	// Inicializar el repositorio GORM con la conexión a la base de datos
	repo := repository.NewPlantillaRepository(dbManager.DB)

	// Crear una instancia del servicio de correo electrónico utilizando SMTP o SES
	var emailService service.EmailService
	emailProvider := viper.GetString("EMAIL_PROVIDER")

	switch emailProvider {
	case "SES":
		awsConfigLoader := &email.RealAWSConfigLoader{}
		emailService = email.NewSESEmailService(awsConfigLoader)
	default:
		var smtpErr error
		emailService, smtpErr = email.NewSMTPEmailService(secretService)
		if smtpErr != nil {
			logs.LogError("Error inicializando el servicio SMTP: %v", smtpErr)
			return nil, nil, smtpErr
		}
	}

	// Crear una instancia del servicio PlantillaService con el servicio de correo electrónico
	plantillaService := service.NewPlantillaService(repo, emailService)

	return plantillaService, dbManager, nil
}

func CleanupApplication(dbManager *connection.DBManager) {
	if dbManager != nil {
		// Cerrar la conexión a la base de datos si existe
		dbManager.CloseDB()
	} else {
		logs.LogError("El DBManager no fue inicializado. No se requiere limpieza.", nil)
	}
}
