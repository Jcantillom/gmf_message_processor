package main

import (
	"context"
	"gmf_message_processor/config"           // Importar el paquete config para inicialización y limpieza
	"gmf_message_processor/internal/handler" // Importar el handler
	"log"
)

func main() {
	// Inicializar la aplicación y obtener el servicio necesario y el DBManager
	plantillaService, dbManager := config.InitApplication()

	// Asegurarse de limpiar los recursos al finalizar
	defer config.CleanupApplication(dbManager)

	// Crear un contexto para la operación de procesamiento de mensajes SQS
	ctx := context.TODO()

	// Procesar mensajes de SQS usando el handler
	log.Println("Procesando mensajes de SQS... 🚀")
	processed, err := handler.ProcessSQSMessages(ctx, plantillaService)
	if err != nil {
		log.Fatalf("Error procesando mensajes de SQS ❌: %v", err)
	}

	// Mostrar el resultado del procesamiento
	log.Printf("Mensajes procesados con éxito ✅: %d", processed)
}
