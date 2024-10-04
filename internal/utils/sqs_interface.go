package utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSAPI es una interfaz que define los métodos que necesitamos del cliente de SQS.
type SQSAPI interface {
	DeleteMessage(
		ctx context.Context,
		input *sqs.DeleteMessageInput,
		opts ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}
