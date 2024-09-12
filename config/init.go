package config

import (
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/service"
	"log"
)

// InitApplication inicializa la configuración, la conexión a la base de datos y crea las instancias de servicios necesarias.
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
	repo := repository.NewGormPlantillaRepository(dbManager.DB)

	// Crear una instancia del servicio PlantillaService
	plantillaService := service.NewPlantillaService(repo, nil)

	return plantillaService, dbManager
}

// CleanupApplication maneja la limpieza de recursos, como cerrar conexiones a la base de datos.
func CleanupApplication(dbManager *connection.DBManager) {
	// Cerrar la conexión a la base de datos
	dbManager.CloseDB()
	log.Println("Recursos limpiados correctamente 🧹")
}
