package repository

import (
	"gmf_message_processor/internal/models"
	"log"

	"gorm.io/gorm"
)

// GormPlantillaRepository implementa el repositorio de Plantilla utilizando GORM.
type GormPlantillaRepository struct {
	DB *gorm.DB
}

func NewPlantillaRepository(db *gorm.DB) *GormPlantillaRepository {
	return &GormPlantillaRepository{DB: db}
}

// CheckPlantillaExists verifica si una plantilla existe en la base de datos y la devuelve.
func (repo *GormPlantillaRepository) CheckPlantillaExists(idPlantilla string) (
	bool, *models.Plantilla, error) {
	var plantilla models.Plantilla
	if err := repo.DB.Where(
		"id_plantilla = ?",
		idPlantilla).First(&plantilla).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, nil
		}
		log.Printf("Error al verificar la existencia de la plantilla con ID %s: %v", idPlantilla, err)
		return false, nil, err
	}
	return true, &plantilla, nil
}
