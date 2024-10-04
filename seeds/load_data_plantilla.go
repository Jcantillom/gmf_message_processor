package seeds

import (
	"context"
	"encoding/json"
	"gmf_message_processor/connection"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"os"
)

func SeedDataPlantilla(ctx context.Context, dbManager *connection.DBManager) {
	// Leer datos del archivo JSON
	plantillas, err := loadPlantillaFromJSON("seeds/data/plantillas.json")
	if err != nil {
		logs.LogError(
			ctx,
			"Error al leer datos de semilla desde el archivo JSON ❌: %v",
			err,
		)
	}

	// Insertar datos de semilla en la base de datos
	for _, plantilla := range plantillas {
		// Verificar que el ID de la plantilla no esté vacío
		if plantilla.IDPlantilla == "" {
			logs.LogError(
				ctx,
				"El ID de la plantilla no puede estar vacío ❌: %v",
				plantilla,
			)
			continue
		}

		// Verificar si la plantilla ya existe en la base de datos
		var existingPlantilla models.Plantilla
		if err := dbManager.DB.Where(
			"id_plantilla = ?", plantilla.IDPlantilla).First(&existingPlantilla).Error; err == nil {
			continue // Saltar si ya existe
		}

		// Insertar la plantilla en la base de datos
		if err := dbManager.DB.Create(&plantilla).Error; err != nil {
			logs.LogError(
				ctx,
				"Error al insertar plantilla en la base de datos ❌: %v",
				err,
			)
		} else {
			logs.LogPlantillaInsertada(ctx, plantilla.IDPlantilla)
		}
		logs.LogDatosSemillaPlantillaInsertados(ctx)
	}

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
