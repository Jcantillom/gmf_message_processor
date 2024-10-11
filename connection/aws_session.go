package connection

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"gmf_message_processor/internal/logs"
)

// NewSession crea una nueva sesión de AWS según el entorno
func NewSession(messageID string) (*session.Session, error) {
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
			Region: aws.String("us-east-1"),
		}
	}

	// Crear una nueva sesión de AWS
	newSession, err := session.NewSession(config)
	if err != nil {
		logs.LogError("Error al crear la sesión de AWS", err, messageID)
		return nil, err
	}

	return newSession, nil
}
