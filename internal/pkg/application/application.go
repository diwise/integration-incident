package application

import (
	"fmt"
	"strings"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type IntegrationIncident interface {
	Start() error
	DeviceStateUpdated(deviceId, deviceState string) error
}

type app struct {
	log              zerolog.Logger
	incidentReporter func(models.Incident) error
	baseUrl          string
	port             string
	previousStates   map[string]string
}

func NewApplication(log zerolog.Logger, incidentReporter func(models.Incident) error, baseUrl, port string) IntegrationIncident {
	prevState := make(map[string]string)

	newApp := &app{
		log:              log,
		incidentReporter: incidentReporter,
		baseUrl:          baseUrl,
		port:             port,
		previousStates:   prevState,
	}

	return newApp
}

func (a *app) Start() error {

	go a.RunPoll(a.log, a.baseUrl, a.incidentReporter)

	return nil
}

func (a *app) RunPoll(log zerolog.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {
	err := GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to start polling for devices: %s", err.Error())
	}

	for {
		log.Info().Msg("Polling for device status ...")
		GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
		time.Sleep(5 * time.Second)
	}
}

func (a *app) DeviceStateUpdated(deviceId, deviceState string) error {
	prefix := fiware.DeviceIDPrefix + "se:servanet:lora:msva:"
	if !strings.HasPrefix(deviceId, prefix) {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	shortId := strings.TrimPrefix(deviceId, prefix)

	if !a.deviceStateHasChanged(shortId, deviceState) {
		return nil
	}

	err := a.createAndSendIncident(shortId, deviceState, a.incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to create and send incident: %s", err.Error())
	}

	a.updateDeviceState(shortId, deviceState)

	return nil
}

func (a *app) updateDeviceState(deviceId, deviceState string) {
	a.previousStates[deviceId] = deviceState
}

func (a *app) deviceStateHasChanged(deviceId, state string) bool {
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
	incident.Description = getDescriptionFromDeviceState(deviceId, state)

	log.Info().Msg("sending incident")
	err := incidentReporter(incident)
	if err != nil {
		log.Err(err).Msg("could not post incident")
		return err
	}

	return nil
}

func getDescriptionFromDeviceState(deviceId, state string) string {

	description, ok := deviceStateDescriptions[state]
	if !ok {
		description = fmt.Sprintf("Okänt Fel: %s", state)
	}

	return fmt.Sprintf("%s - %s", deviceId, description)
}

var deviceStateDescriptions map[string]string = map[string]string{
	"0":   "Inga Fel",
	"4":   "Låg Batterinivå",
	"8":   "Permanent Fel",
	"16":  "Temporärt Fel Tomt Rör",
	"18":  "Låg Betterinivå Permanent Fel",
	"20":  "Tomt Rör och Temporärt Fel Låg Batterinivå",
	"24":  "Permanent Fel och Temporärt Fel Tomt Rör",
	"34":  "Permanent Fel Låg Batterinivå och Temporärt Fel Tomt Rör",
	"48":  "Temporärt Fel Läckage",
	"52":  "Läckage och Temporärt Fel Låg Batterinivå",
	"56":  "Permanent Fel och Temporärt Fel Läckage",
	"66":  "Permanent Fel Låg Batterinivå och Temporärt Fel Läckage",
	"112": "Temporärt Fel Backflöde",
	"116": "Backflöde och Temporärt Fel Låg Batterinivå",
	"120": "Permanent Fel och Temporärt Fel Backflöde",
	"130": "Permanent Fel Låg Batterinivå och Temporärt Fel Backflöde",
	"144": "Temporärt Fel Is eller Frys Varning",
	"148": "Is eller Frys Varning och Temporärt Fel Låg Batterinivå",
	"152": "Permanent Fel och Temporärt Fel Is eller Frys Varning",
	"156": "Permanent Fel Låg Batterinivå och Temporärt Fel Is eller Frys Varning",
	"176": "Temporärt Fel Spricka eller Öppet Rör",
	"180": "Spricka eller Öppet RÖr och Temporärt Fel Låg Batterinivå",
	"184": "Permanent Fel och Temporärt Fel Spricka eller Öppet Rör",
	"188": "Permanent Fel Låg Batterinivå och Temporärt Fel Spricka eller Öppet Rör",
}
