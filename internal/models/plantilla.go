package models

import (
	"fmt"
	"os"
	"time"
)

// Plantilla representa la estructura del modelo de Plantilla.
type Plantilla struct {
	IDPlantilla  string    `json:"IDPlantilla" gorm:"type:char(5);not null;primaryKey"`
	Asunto       string    `json:"Asunto" gorm:"type:varchar(255);not null"`
	Cuerpo       string    `json:"Cuerpo" gorm:"type:text;not null"`
	Remitente    string    `json:"Remitente" gorm:"type:varchar(100);not null"`
	Destinatario string    `json:"Destinatario" gorm:"type:varchar(1000)"`
	Adjunto      bool      `json:"Adjunto" gorm:"type:boolean;not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName devuelve el nombre de la tabla para el modelo Plantilla.
func (Plantilla) TableName() string {
	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		return "cgd_correos_plantillas"
	}
	return fmt.Sprintf("%s.cgd_correos_plantillas", schema)
}
