package service

import (
	"context"
	"errors"
	"gmf_message_processor/internal/email"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/repository"
	"gmf_message_processor/internal/utils"
	"log"
)

// HandlePlantilla maneja la lógica de negocio para la plantilla.
func HandlePlantilla(ctx context.Context, msg *models.SQSMessage) error {
	// Verificar si la plantilla existe en la base de datos
	exists, plantilla, err := repository.CheckPlantillaExists(msg.IDPlantilla)
	if err != nil {
		return err
	}
	if !exists {
		log.Printf("La plantilla con ID %s no existe.", msg.IDPlantilla)
		return errors.New("la plantilla no existe en la base de datos")
	}

	// Verificar que haya al menos un conjunto de parámetros en el array
	if len(msg.Parametro) == 0 {
		log.Printf("No se proporcionaron parámetros para la plantilla con ID %s.", msg.IDPlantilla)
		return errors.New("no se proporcionaron parámetros para la plantilla")
	}

	// Usar el primer conjunto de parámetros
	params := msg.Parametro[0]

	// Mapear los parámetros a un mapa de strings
	placeholders := map[string]string{
		"nombre_archivo":      params.NombreArchivo,
		"plataforma_origen":   params.PlataformaOrigen,
		"fecha_recepcion":     params.FechaRecepcion,
		"hora_recepcion":      params.HoraRecepcion,
		"codigo_rechazo":      params.CodigoRechazo,
		"descripcion_rechazo": params.DescripcionRechazo,
		"detalle_rechazo":     params.DetalleRechazo,
	}

	// Reemplazar los placeholders en el cuerpo de la plantilla
	plantilla.Cuerpo = utils.ReplacePlaceholders(plantilla.Cuerpo, placeholders)

	// Si la plantilla existe, enviar el correo electrónico
	log.Printf("Plantilla con ID %s encontrada. Enviando correo electrónico...", msg.IDPlantilla)
	err = email.SendEmail(plantilla.Remitente, plantilla.Destinatario, plantilla.Asunto, plantilla.Cuerpo)
	if err != nil {
		log.Printf("Error enviando el correo para la plantilla con ID %s: %v", msg.IDPlantilla, err)
		return err
	}

	log.Printf("Correo electrónico enviado exitosamente para IDPlantilla: %s", msg.IDPlantilla)
	return nil
}
