package service

import (
	"context"
	"errors"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
)

type IPlantillaService interface {
	HandlePlantilla(ctx context.Context, msg *models.SQSMessage) error
}

// EmailService define la interfaz para el servicio de correo electrónico.
type EmailService interface {
	SendEmail(remitente, destinatarios, asunto, cuerpo string) error
}

// PlantillaRepository define la interfaz para el repositorio de Plantilla.
type PlantillaRepository interface {
	CheckPlantillaExists(idPlantilla string) (bool, *models.Plantilla, error)
}

// PlantillaService define el servicio que maneja la lógica de negocio para Plantilla.
type PlantillaService struct {
	repo         PlantillaRepository
	emailService EmailService
}

// NewPlantillaService crea una nueva instancia de PlantillaService.
func NewPlantillaService(repo PlantillaRepository, emailService EmailService) *PlantillaService {
	return &PlantillaService{
		repo:         repo,
		emailService: emailService,
	}
}

func (s *PlantillaService) HandlePlantilla(ctx context.Context, msg *models.SQSMessage) error {
	// Verificar si la plantilla existe en la base de datos
	exists, plantilla, err := s.repo.CheckPlantillaExists(msg.IDPlantilla)
	if err != nil {
		logs.LogError(
			ctx,
			"Error al verificar si la plantilla con ID %s existe en la base de datos: %v",
			msg.IDPlantilla,
			err,
		)
		return err
	}
	if !exists {
		logs.LogPlantillaNoEncontrada(ctx, msg.IDPlantilla)
		return errors.New("la plantilla no existe en la base de datos")
	}

	// Verificar que haya al menos un conjunto de parámetros en el array
	if len(msg.Parametro) == 0 {
		logs.LogParametrosNoProporcionados(ctx, msg.IDPlantilla)
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

	// Si la plantilla existe, enviar el correo electrónico usando el servicio de correo
	logs.LogPlantillaEncontrada(ctx, msg.IDPlantilla)
	err = s.emailService.SendEmail(plantilla.Remitente, plantilla.Destinatario, plantilla.Asunto, plantilla.Cuerpo)
	if err != nil {
		logs.LogErrorEnvioCorreo(ctx, msg.IDPlantilla, err)
		return err
	}

	logs.LogCorreoEnviado(ctx, msg.IDPlantilla)
	return nil
}
