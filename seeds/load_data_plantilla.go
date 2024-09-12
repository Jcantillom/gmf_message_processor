package seeds

import (
	"encoding/json"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/models"
	"log"
	"os"
)

// SeedDataPlantilla inserta datos de semilla en la base de datos utilizando una instancia de DBManager.
func SeedDataPlantilla(dbManager *connection.DBManager) {
	// Leer datos del archivo JSON
	plantillas, err := loadPlantillaFromJSON("seeds/data/plantillas.json")
	if err != nil {
		log.Fatalf("Error al cargar datos de plantilla desde el archivo JSON ❌: %v", err)
	}

	// Insertar datos de semilla en la base de datos
	for _, plantilla := range plantillas {
		// Verificar que el ID de la plantilla no esté vacío
		if plantilla.IDPlantilla == "" {
			log.Println("El campo IDPlantilla está vacío. Saltando esta plantilla. ⚠️")
			continue
		}

		// Verificar si la plantilla ya existe en la base de datos
		var existingPlantilla models.Plantilla
		if err := dbManager.DB.Where(
			"id_plantilla = ?", plantilla.IDPlantilla).First(&existingPlantilla).Error; err == nil {
			log.Printf("La plantilla con ID %s ya existe en la base de datos. ⚠️", plantilla.IDPlantilla)
			continue // Saltar si ya existe
		}

		// Insertar la plantilla en la base de datos
		if err := dbManager.DB.Create(&plantilla).Error; err != nil {
			log.Fatalf("Error al insertar datos de semilla en la base de datos ❌: %v", err)
		} else {
			log.Printf(
				"Plantilla con ID %s insertada correctamente en la base de datos. ✅", plantilla.IDPlantilla)
		}
	}

	log.Println("Datos de semilla insertados correctamente en la base de datos 🌱")
}

// loadPlantillaFromJSON carga los datos de plantilla desde un archivo JSON
func loadPlantillaFromJSON(filePath string) ([]models.Plantilla, error) {
	// Abrir el archivo JSON
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decodificar el archivo JSON
	var plantillas []models.Plantilla
	if err := json.NewDecoder(file).Decode(&plantillas); err != nil {
		return nil, err
	}

	return plantillas, nil
}
