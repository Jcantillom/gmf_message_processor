package config

import (
	"context"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
	"gmf_message_processor/seeds"
	"log"
)

var ctx = context.TODO()

// Inicializa la configuración, la conexión a la base de datos y crea las instancias de servicios necesarias.
func InitApplication() (*service.PlantillaService, *connection.DBManager) {
	// Inicializar el ConfigManager y cargar configuración
	configManager := NewConfigManager()
	configManager.InitConfig()

	// Inicializar el DBManager y abrir la conexión a la base de datos
	dbManager := connection.NewDBManager()
	if err := dbManager.InitDB(); err != nil {
		log.Fatalf("Error inicializando la base de datos: %v", err)
	}

	// Inicializar el repositorio GORM con la conexión a la base de datos
	repo := repository.NewPlantillaRepository(dbManager.DB)

	// Crear una instancia del servicio de correo electrónico utilizando SMTP
	emailService := email.NewSMTPEmailService()

	// Crear una instancia del servicio PlantillaService con el servicio de correo electrónico
	plantillaService := service.NewPlantillaService(repo, emailService)

	// Insertar datos de semilla en la base de datos
	seeds.SeedDataPlantilla(nil, dbManager)

	return plantillaService, dbManager
}

// maneja la limpieza de recursos, como cerrar conexiones a la base de datos.
func CleanupApplication(dbManager connection.DBManagerInterface) {
	// Cerrar la conexión a la base de datos
	dbManager.CloseDB()
	logs.LogInfo(ctx, "Recursos limpiados correctamente 🧹 ", nil)

}
