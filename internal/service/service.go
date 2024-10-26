package service

import (
	"context"
	"errors"
	"fmt"
	"gmf_message_processor/internal/logs"
	"gmf_message_processor/internal/models"
	"gmf_message_processor/internal/utils"
)

type IPlantillaService interface {
	HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error
}

// EmailService define la interfaz para el servicio de correo electrónico.
type EmailService interface {
	SendEmail(
		remitente,
		destinatarios,
		asunto,
		cuerpo,
		imagePath,
		imageName,
		messageID string) error
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

func (s *PlantillaService) HandlePlantilla(ctx context.Context, msg *models.SQSMessage, messageID string) error {
	// Verificar si la plantilla existe en la base de datos
	exists, plantilla, err := s.repo.CheckPlantillaExists(msg.IDPlantilla)
	if err != nil {
		logs.LogError(
			fmt.Sprintf("Error al verificar si la plantilla con ID %s existe en la base de datos", msg.IDPlantilla),
			err,
			messageID,
		)
		return err
	}
	if !exists {
		logs.LogError(fmt.Sprintf("La plantilla con ID %s no existe en la base de datos", msg.IDPlantilla), nil, messageID)
		return errors.New("la plantilla no existe en la base de datos")
	}
	// leer la imagen desde el archivo y adjuntarla al correo
	imagePath := "images/Casitadavivienda.png"
	imageName := "logo.png"

	// Verificar que haya al menos un conjunto de parámetros en el array
	if len(msg.Parametro) == 0 {
		logs.LogInfo(
			"No se proporcionaron parámetros para la plantilla. Se utilizarán valores predeterminados.",
			messageID,
		)

		// Aquí puedes usar valores predeterminados o simplemente continuar sin los parámetros.
		placeholders := map[string]string{}

		// Reemplazar placeholders en el cuerpo de la plantilla
		plantilla.Cuerpo = utils.ReplacePlaceholders(plantilla.Cuerpo, placeholders)

		// Continuar con el envío de correo aunque no haya parámetros
		err = s.emailService.SendEmail(
			plantilla.Remitente,
			plantilla.Destinatario,
			plantilla.Asunto,
			plantilla.Cuerpo,
			imagePath,
			imageName,
			messageID)
		if err != nil {
			logs.LogError("Error al enviar el correo electrónico", err, messageID)
			return err
		}

		logs.LogInfo("Correo electrónico enviado sin parámetros", messageID)
		return nil
	}

	// Iterar sobre todos los parámetros y mapear los valores
	placeholders := map[string]string{}
	for _, param := range msg.Parametro {
		placeholders["&"+param.Nombre] = param.Valor
	}

	// Reemplazar los placeholders en el cuerpo de la plantilla
	plantilla.Cuerpo = utils.ReplacePlaceholders(plantilla.Cuerpo, placeholders)

	// Enviar el correo electrónico usando el servicio de correo
	err = s.emailService.SendEmail(
		plantilla.Remitente,
		plantilla.Destinatario,
		plantilla.Asunto,
		plantilla.Cuerpo,
		imagePath,
		imageName,
		messageID,
	)
	if err != nil {
		logs.LogError("Error al enviar el correo electrónico", err, messageID)
		panic(err)
	}

	return nil
}
