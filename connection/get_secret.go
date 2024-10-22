package connection

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"gmf_message_processor/internal/logs"
)

// SecretService interface para obtener secretos
type SecretService interface {
	GetSecret(secretName string, messageID string) (*SecretData, error)
}

// SecretServiceImpl es la implementación de SecretService
type SecretServiceImpl struct {
	secretsmanager secretsmanageriface.SecretsManagerAPI
}

// SecretData estructura para almacenar los datos del secreto
type SecretData struct {
	Username string `json:"USERNAME"`
	Password string `json:"PASSWORD"`
}

// NewSecretService crea una nueva instancia de SecretService
func NewSecretService(sess *session.Session) SecretService {
	return &SecretServiceImpl{
		secretsmanager: secretsmanager.New(sess),
	}
}

// GetSecret obtiene el secreto de AWS Secrets Manager o LocalStack
func (s *SecretServiceImpl) GetSecret(secretName string, messageID string) (*SecretData, error) {
	if secretName == "" {
		logs.LogError("El nombre del secreto no puede estar vacío", nil, messageID)
		return nil, errors.New("el nombre del secreto no puede estar vacío")
	}

	// El cliente inyectado de Secrets Manager se usa en lugar de crear uno nuevo aquí
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := s.secretsmanager.GetSecretValue(input)
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

	if result.SecretString == nil {
		logs.LogError("El secreto es nulo", nil, messageID)
		return nil, fmt.Errorf("el secreto '%s' es nulo", secretName)
	}

	var secretData SecretData
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		logs.LogError("Error al deserializar el secreto", err, messageID)
		return nil, fmt.Errorf("error al deserializar el secreto: %w", err)
	}

	return &secretData, nil
}
