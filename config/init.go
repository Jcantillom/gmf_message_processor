package config

import (
	"context"
	"fmt"
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
	"os"
)

type AppContext struct {
	PlantillaService service.IPlantillaService
	DBManager        connection.DBManagerInterface
	SQSHandler       handler.SQSHandlerInterface
	Logger           logs.LogInterface
	SQSClient        internalAws.SQSAPI
	Utils            *utils.Utils
}

func logDatabaseConnectionEstablished(messageID string) {
	logs.Logger.LogDebug("Conexión a la base de datos establecida", messageID)
}

func initializeRepository(dbManager connection.DBManagerInterface) *repository.GormPlantillaRepository {
	return repository.NewPlantillaRepository(dbManager.GetDB())
}

// InitApplication inicializa la aplicación y devuelve un AppContext con los servicios y configuraciones necesarios.
func InitApplication(
	messageID string,
	secretService connection.SecretService,
	dbManager connection.DBManagerInterface) (*AppContext, error) {
	// Inicializar el ConfigManager y cargar configuración
	configManager := NewConfigManager(&logs.LoggerAdapter{})
	configManager.InitConfig(messageID)

	// Validar y obtener secretos necesarios
	_, err := getSecret(secretService, "SECRETS_DB", messageID)
	if err != nil {
		return nil, err
	}

	_, err = getSecret(secretService, "SECRETS_SMTP", messageID)
	if err != nil {
		return nil, err
	}

	// Inicializar el Logger
	logger := &logs.LoggerAdapter{}
	utilsImpl := &utils.Utils{}

	// Inicializar el DBManager y abrir la conexión a la base de datos
	err = dbManager.InitDB(messageID)
	if err != nil {
		logs.LogError("Error inicializando la base de datos", err, messageID)
		return nil, err
	}

	logDatabaseConnectionEstablished(messageID)

	// Inicializar el repositorio GORM con la conexión a la base de datos
	repo := repository.NewPlantillaRepository(dbManager.GetDB())

	// Crear una instancia del servicio de correo electrónico utilizando SMTP
	emailService, smtpErr := email.NewSMTPEmailService(secretService, messageID)
	if smtpErr != nil {
		logs.LogError("Error inicializando el servicio SMTP", smtpErr, messageID)
		return nil, smtpErr
	}

	// Crear una instancia del servicio PlantillaService
	plantillaService := service.NewPlantillaService(repo, emailService)

	// Inicializar el cliente SQS
	sqsClient, err := initializeSQSClient(messageID)
	if err != nil {
		return nil, err
	}
	// Obtener la URL de la cola SQS desde la configuración
	queueURL := viper.GetString("SQS_QUEUE_URL")

	// Crear un nuevo manejador de SQS
	sqsHandler := handler.NewSQSHandler(
		plantillaService,
		sqsClient,
		utilsImpl,
		logger,
		queueURL,
	)

	return &AppContext{
		PlantillaService: plantillaService,
		DBManager:        dbManager,
		SQSHandler:       sqsHandler,
		Logger:           logger,
		SQSClient:        sqsClient,
		Utils:            utilsImpl,
	}, nil
}

// getSecret obtiene un secreto validando que esté configurado en las variables de entorno
func getSecret(
	secretService connection.SecretService,
	envVar string,
	messageID string) (*connection.SecretData, error) {
	secretName := os.Getenv(envVar)
	if secretName == "" {
		logs.LogError(fmt.Sprintf("La variable %s no está configurada", envVar), nil, messageID)
		return nil, fmt.Errorf("la variable %s no está configurada", envVar)
	}

	secret, err := secretService.GetSecret(secretName, messageID)
	if err != nil {
		logs.LogError(fmt.Sprintf("Error al obtener %s", envVar), err, messageID)
		return nil, err
	}

	return secret, nil
}

// initializeSQSClient inicializa el cliente de SQS, permitiendo la configuración del endpoint.
func initializeSQSClient(messageID string) (internalAws.SQSAPI, error) {
	// Verificar si estamos utilizando LocalStack o AWS real
	endpoint := viper.GetString("SQS_ENDPOINT")
	region := viper.GetString("AWS_REGION")

	if region == "" {
		region = "us-east-1" // Región por defecto
	}

	loadConfigFunc := func(ctx context.Context, optFns ...func(*awsConfig.LoadOptions) error) (aws.Config, error) {
		options := []func(*awsConfig.LoadOptions) error{
			awsConfig.WithRegion(region),
		}

		// Si se proporciona un endpoint (por ejemplo, para LocalStack), lo agregamos
		if endpoint != "" {
			options = append(options, awsConfig.WithEndpointResolver(
				aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				}),
			))
		}

		return awsConfig.LoadDefaultConfig(ctx, options...)
	}

	// Inicializar el cliente SQS
	sqsClient, err := internalAws.NewSQSClient(viper.GetString("SQS_QUEUE_URL"), loadConfigFunc)
	if err != nil {
		logs.LogError("Error inicializando SQS Client", err, messageID)
		return nil, err
	}

	return sqsClient, nil
}

// CleanupApplication realiza la limpieza del DBManager
func CleanupApplication(dbManager connection.DBManagerInterface, messageID string) {
	if dbManager != nil {
		dbManager.CloseDB(messageID)
	} else {
		logs.LogError("El DBManager no fue inicializado. No se requiere limpieza.", nil, messageID)
	}
}
