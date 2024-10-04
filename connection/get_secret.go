package connection

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"gmf_message_processor/internal/logs"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// SecretData estructura para almacenar los datos del secreto
type SecretData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// getSecret obtiene el secreto de AWS Secrets Manager o LocalStack
func getSecret(secretName string) (*SecretData, error) {
	var sess *session.Session
	var err error

	environment := os.Getenv("APP_ENV")

	if environment == "local" {
		// Configuración para LocalStack
		logs.LogInfo(nil, "Usando LocalStack")
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String("us-east-1"),
			Endpoint:    aws.String("http://localhost:4566"),                  // Apunta a LocalStack
			Credentials: credentials.NewStaticCredentials("test", "test", ""), // Credenciales de prueba
		})
	} else {
		// Configuración para AWS
		sess, err = session.NewSession(&aws.Config{
			Region: aws.String("us-east-1"), // Cambia a tu región real
		})
	}

	if err != nil {
		return nil, fmt.Errorf("error al crear la sesión de AWS: %w", err)
	}

	svc := secretsmanager.New(sess)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return nil, fmt.Errorf("error al obtener el secreto: %w", err)
	}

	var secretData SecretData
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		return nil, fmt.Errorf("error al deserializar el secreto: %w", err)
	}

	return &secretData, nil
}
