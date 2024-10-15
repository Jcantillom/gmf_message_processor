package config

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/viper"
	"gmf_message_processor/connection"
	internalAws "gmf_message_processor/internal/aws"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/handler"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
	"gmf_message_processor/internal/utils"
)

type AppContext struct {
	PlantillaService *service.PlantillaService
	DBManager        *connection.DBManager
	SQSHandler       *handler.SQSHandler
	Logger           *logs.LoggerAdapter
	SQSClient        *internalAws.SQSClient
	Utils            *utils.Utils
}

// InitApplication Inicializa la aplicación y devuelve un AppContext con los servicios y configuraciones necesarios.
func InitApplication(messageID string) (*AppContext, error) {
	// Inicializar el ConfigManager y cargar configuración
	configManager := NewConfigManager()
	configManager.InitConfig(messageID)

	// Inicializar el Logger
	logger := &logs.LoggerAdapter{}

	// Crear una Instancia de Utils (que implementa UtilsInterface)
	utilsImpl := &utils.Utils{}

	// Inicializar el Servicio de Secretos
	secretService := connection.NewSecretService()

	// Inicializar el DBManager y abrir la conexión a la base de datos
	dbManager := connection.NewDBManager(secretService)
	err := dbManager.InitDB(messageID)
	if err != nil {
		logs.LogError("Error inicializando la base de datos", err, messageID)
		return nil, err
	}
	logs.LogDebug("Conexión a la base de datos establecida", messageID)

	// Inicializar el repositorio GORM con la conexión a la base de datos
	repo := repository.NewPlantillaRepository(dbManager.DB)

	// Crear una instancia del servicio de correo electrónico utilizando SMTP
	emailService, smtpErr := email.NewSMTPEmailService(secretService, messageID) // Pasar messageID
	if smtpErr != nil {
		logs.LogError("Error inicializando el servicio SMTP", smtpErr, messageID)
		return nil, smtpErr
	}
	// Crear una instancia del servicio PlantillaService con el servicio de correo electrónico
	plantillaService := service.NewPlantillaService(repo, emailService)

	// Función para cargar la configuración de AWS
	loadConfigFunc := func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (
		aws.Config,
		error,
	) {
		return awsConfig.LoadDefaultConfig(ctx, optFns...)
	}

	// Inicializar SQSClient
	sqsClient, err := internalAws.NewSQSClient(viper.GetString("SQS_QUEUE_URL"), loadConfigFunc)
	if err != nil {
		logs.LogError("Error inicializando SQS Client", err, messageID)
		return nil, err
	}

	// Crear un nuevo manejador de SQS
	sqsHandler := handler.NewSQSHandler(plantillaService, sqsClient, utilsImpl, logger)

	return &AppContext{
		PlantillaService: plantillaService,
		DBManager:        dbManager,
		SQSHandler:       sqsHandler,
		Logger:           logger,
		SQSClient:        sqsClient,
		Utils:            utilsImpl,
	}, nil
}

func CleanupApplication(dbManager *connection.DBManager, messageID string) {
	if dbManager != nil {
		dbManager.CloseDB(messageID)
	} else {
		logs.LogError("El DBManager no fue inicializado. No se requiere limpieza.", nil, messageID)
	}
}
