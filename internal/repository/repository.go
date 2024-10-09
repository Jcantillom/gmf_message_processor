package repository

import (
	"errors"
	"gmf_message_processor/internal/models"
	"gorm.io/gorm"
)

// DBInterface define las operaciones mínimas de la base de datos que necesitamos.
type DBInterface interface {
	Where(query interface{}, args ...interface{}) *gorm.DB
	First(dest interface{}, conds ...interface{}) *gorm.DB
}

// GormPlantillaRepository implementa el repositorio de Plantilla utilizando GORM.
type GormPlantillaRepository struct {
	DB DBInterface
}

func NewPlantillaRepository(db DBInterface) *GormPlantillaRepository {
	return &GormPlantillaRepository{DB: db}
}

// CheckPlantillaExists verifica si una plantilla existe en la base de datos y la devuelve.
func (repo *GormPlantillaRepository) CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error) {
	var plantilla models.Plantilla

	if err := repo.DB.Where("id_plantilla = ?", idPlantilla).First(&plantilla).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &plantilla, nil
}
