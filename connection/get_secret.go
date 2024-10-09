package connection

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"gmf_message_processor/internal/logs"
)

// SecretService interface para obtener secretos
type SecretService interface {
	GetSecret(secretName string) (*SecretData, error)
}

// SecretServiceImpl es la implementación de SecretService
type SecretServiceImpl struct{}

// SecretData estructura para almacenar los datos del secreto
type SecretData struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
}

// NewSecretService crea una nueva instancia de SecretService
func NewSecretService() SecretService {
	return &SecretServiceImpl{}
}

// NewSession crea una nueva sesión de AWS según el entorno
func NewSession() (*session.Session, error) {
	environment := os.Getenv("APP_ENV")
	var config *aws.Config

	if environment == "local" {
		config = &aws.Config{
			Region:      aws.String("us-east-1"),
			Endpoint:    aws.String("http://localhost:4566"),
			Credentials: credentials.NewStaticCredentials("test", "test", ""),
		}
	} else {
		config = &aws.Config{
			Region: aws.String("us-east-1"), // Cambia a tu región real
		}
	}

	return session.NewSession(config)
}

// GetSecret obtiene el secreto de AWS Secrets Manager o LocalStack
func (s *SecretServiceImpl) GetSecret(secretName string) (*SecretData, error) {
	// Crear sesión de AWS
	sess, err := NewSession()
	if err != nil {
		logs.LogError("Error al crear la sesión de AWS: %v", err)
		return nil, fmt.Errorf("error al crear la sesión de AWS: %w", err)
	}

	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	// Intentar obtener el secreto
	result, err := svc.GetSecretValue(input)
	if err != nil {
		// Verificar si es un error de secreto no encontrado
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
			logs.LogWarn("Secreto no encontrado en AWS Secrets Manager", "secretName", secretName)
			return nil, fmt.Errorf("secreto no encontrado: %s", secretName)
		}
		logs.LogError("Error al obtener el secreto: %v", err)
		return nil, fmt.Errorf("error al obtener el secreto: %w", err)
	}

	// Validar que el SecretString no sea nil antes de desreferenciarlo
	if result.SecretString == nil {
		logs.LogError("El secreto es nulo", err)
		return nil, fmt.Errorf("el secreto '%s' es nulo", secretName)
	}

	// Deserializar el contenido del secreto
	var secretData SecretData
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		logs.LogError("Error al deserializar el secreto", err)
		return nil, fmt.Errorf("error al deserializar el secreto: %w", err)
	}

	// Devolver el secreto si todo fue exitoso
	return &secretData, nil
}
