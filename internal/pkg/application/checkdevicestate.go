package application

import (
	"fmt"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/rs/zerolog/log"
)

func (a *app) checkDeviceExistsAndPreviousDeviceState(deviceId, state string) bool {
	_, exists := a.previousStates[deviceId]

	if !exists {
		log.Info().Msg("device does not exist, saving state...")
		a.previousStates[deviceId] = state
		return false
	}

	if a.previousStates[deviceId] != state {
		log.Info().Msg("device state has changed")
		return true
	}

	log.Info().Msg("device state has not changed")

	return false
}

func (a *app) createAndSendIncident(deviceId, state string, incidentReporter func(models.Incident) error) error {
	const watermeterCategory int = 16
	incident := models.Incident{}

	incident.PersonId = "diwise"
	incident.Category = watermeterCategory
	incident.Description = fmt.Sprintf("%s - %s", deviceId, state)

	log.Info().Msg("sending incident")
	err := incidentReporter(incident)
	if err != nil {
		log.Err(err).Msg("could not post incident")
		return err
	}

	return nil
}
