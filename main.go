package main

import (
	"github.com/jcantillom/gmf_message_processor/config"
	"github.com/jcantillom/gmf_message_processor/internal/sqs"
)

func main() {
	appConfig, awsConfig := config.LoadConfig()
	sqsClient := sqs.NewSQSClient(appConfig, awsConfig)

	// Consultar mensajes de la cola
	sqsClient.ReceiveMessages()
}
