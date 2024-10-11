package connection

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"gmf_message_processor/internal/logs"
)

// SecretService interface para obtener secretos
type SecretService interface {
	GetSecret(secretName string, messageID string) (*SecretData, error)
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

// GetSecret obtiene el secreto de AWS Secrets Manager o LocalStack
func (s *SecretServiceImpl) GetSecret(secretName string, messageID string) (*SecretData, error) {
	// Crear sesión de AWS
	sess, err := NewSession(messageID)
	if err != nil {
		logs.LogError("Error al crear la sesión de AWS", err, messageID)
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
		var awsErr awserr.Error
		if errors.As(err, &awsErr) && awsErr.Code() == secretsmanager.ErrCodeResourceNotFoundException {
			logs.LogError("Secreto no encontrado en AWS Secrets Manager", err, messageID)
			return nil, fmt.Errorf("secreto no encontrado: %s", secretName)
		}
		logs.LogError("Error al obtener el secreto", err, messageID)
		return nil, fmt.Errorf("error al obtener el secreto: %w", err)
	}

	// Validar que el SecretString no sea nil antes de desreferenciarlo
	if result.SecretString == nil {
		logs.LogError("El secreto es nulo", nil, messageID)
		return nil, fmt.Errorf("el secreto '%s' es nulo", secretName)
	}

	// Deserializar el contenido del secreto
	var secretData SecretData
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		logs.LogError("Error al deserializar el secreto", err, messageID)
		return nil, fmt.Errorf("error al deserializar el secreto: %w", err)
	}

	return &secretData, nil
}
