package application

import (
	"fmt"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/rs/zerolog/log"
)

var previousState map[string]string = make(map[string]string)

func checkPreviousDeviceState(deviceId, state string) bool {
	_, exists := previousState[deviceId]

	if !exists {
		previousState[deviceId] = state
		return false
	}

	if previousState[deviceId] != state {
		return true
	}

	return false
}

func createAndSendIncident(deviceId, state string, incidentReporter func(models.Incident) error) error {
	const watermeterCategory int = 16
	incident := models.Incident{}

	incident.PersonId = "diwise"
	incident.Category = watermeterCategory
	incident.Description = fmt.Sprintf("%s - %s", deviceId, state)

	err := incidentReporter(incident)
	if err != nil {
		log.Err(err).Msg("could not post incident")
		return err
	}

	return nil
}
