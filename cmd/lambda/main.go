package main

import (
	"context"
	"gmf_message_processor/config"
	"gmf_message_processor/internal/handler"
	"log"
)

func main() {
	// Inicializar configuración
	config.InitConfig()

	// Procesar mensajes de SQS
	ctx := context.TODO()
	log.Println("Procesando mensajes de SQS... 🚀")
	process, err := handler.ProcessSQSMessages(ctx)
	if err != nil {
		log.Fatalf("Error procesando mensajes de SQS ❌: %v", err)
	}

	log.Printf("Mensajes procesados con éxito ✅: %d", process)
}
