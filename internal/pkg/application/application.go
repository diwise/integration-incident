package application

import (
	"fmt"
	"strings"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//go:generate moq -rm -out application_mock.go . IntegrationIncident

type IntegrationIncident interface {
	DeviceStateUpdated(deviceId, deviceState string) error
	LifebuoyValueUpdated(deviceId, deviceValue string) error
}

type app struct {
	log              zerolog.Logger
	incidentReporter func(models.Incident) error
	baseUrl          string
	port             string
	previousStates   map[string]string
	previousValues   map[string]string
}

func NewApplication(log zerolog.Logger, incidentReporter func(models.Incident) error, baseUrl, port string) IntegrationIncident {

	newApp := &app{
		log:              log,
		incidentReporter: incidentReporter,
		baseUrl:          baseUrl,
		port:             port,
		previousStates:   make(map[string]string),
		previousValues:   make(map[string]string),
	}

	return newApp
}

func (a *app) DeviceStateUpdated(deviceId, deviceState string) error {
	prefix := fiware.DeviceIDPrefix + "se:servanet:lora:msva:"
	if !strings.HasPrefix(deviceId, prefix) {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	shortId := strings.TrimPrefix(deviceId, prefix)

	if !a.checkIfDeviceStateHasChanged(shortId, deviceState) {
		return nil
	}

	const stateNoError string = "0"

	if deviceState != stateNoError {
		const watermeterCategory int = 17
		incident := models.NewIncident(watermeterCategory, getDescriptionFromDeviceState(shortId, deviceState)).AtLocation(62.388178, 17.315090)

		err := a.incidentReporter(*incident)
		if err != nil {
			log.Err(err).Msg("could not post incident")
			return err
		}
	}

	a.updateDeviceState(shortId, deviceState)

	return nil
}

func (a *app) LifebuoyValueUpdated(deviceId, deviceValue string) error {
	if !strings.HasPrefix(deviceId, diwise.LifebuoyIDPrefix) {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	shortId := strings.TrimPrefix(deviceId, diwise.LifebuoyIDPrefix)
	valueChanged := a.checkIfDeviceValueHasChanged(shortId, deviceValue)

	if !valueChanged {
		return nil
	}

	if deviceValue == "off" {
		log.Info().Msgf("state changed to \"off\" on device: %s", shortId)

		const lifebuoyCategory int = 15
		incident := models.NewIncident(lifebuoyCategory, "Livboj kan ha flyttats eller utsatts f??r ??verkan.")

		lifebuoy, err := getLifebuoyFromContextBroker(a.log, a.baseUrl, deviceId)

		if err == nil {
			point := lifebuoy.Location.GetAsPoint()
			incident = incident.AtLocation(point.Latitude(), point.Longitude())
		}

		err = a.incidentReporter(*incident)
		if err != nil {
			log.Err(err).Msg("could not post incident")
			return err
		}
	}

	a.updateDeviceValue(shortId, deviceValue)

	return nil
}

func (a *app) updateDeviceState(deviceId, deviceState string) {
	a.previousStates[deviceId] = deviceState
}

func (a *app) updateDeviceValue(deviceId, deviceValue string) {
	a.previousValues[deviceId] = deviceValue
}

func (a *app) checkIfDeviceStateHasChanged(deviceId, state string) bool {
	storedState, exists := a.previousStates[deviceId]

	if !exists {
		log.Info().Msgf("device %s does not exist, saving state...", deviceId)
		a.previousStates[deviceId] = state
		return false
	}

	if storedState != state {
		log.Info().Msgf("device %s state has changed from %s to %s", deviceId, storedState, state)
		return true
	}

	return false
}

func (a *app) checkIfDeviceValueHasChanged(deviceId, value string) bool {
	storedValue, exists := a.previousValues[deviceId]

	if !exists {
		log.Info().Msgf("device %s does not exist, saving value...", deviceId)
		a.previousValues[deviceId] = value
		return false
	}

	if storedValue != value {
		log.Info().Msgf("device %s value has changed to %s", deviceId, value)
		return true
	}

	return false
}

func getDescriptionFromDeviceState(deviceId, state string) string {

	description, ok := deviceStateDescriptions[state]
	if !ok {
		description = fmt.Sprintf("Ok??nt Fel: %s", state)
	}

	return fmt.Sprintf("%s - %s", deviceId, description)
}

var deviceStateDescriptions map[string]string = map[string]string{
	"0":   "Inga Fel",
	"4":   "L??g Batteriniv??",
	"8":   "Permanent Fel",
	"16":  "Tempor??rt Fel Tomt R??r",
	"18":  "L??g Betteriniv?? Permanent Fel",
	"20":  "Tomt R??r och Tempor??rt Fel L??g Batteriniv??",
	"24":  "Permanent Fel och Tempor??rt Fel Tomt R??r",
	"34":  "Permanent Fel L??g Batteriniv?? och Tempor??rt Fel Tomt R??r",
	"48":  "Tempor??rt Fel L??ckage",
	"52":  "L??ckage och Tempor??rt Fel L??g Batteriniv??",
	"56":  "Permanent Fel och Tempor??rt Fel L??ckage",
	"66":  "Permanent Fel L??g Batteriniv?? och Tempor??rt Fel L??ckage",
	"112": "Tempor??rt Fel Backfl??de",
	"116": "Backfl??de och Tempor??rt Fel L??g Batteriniv??",
	"120": "Permanent Fel och Tempor??rt Fel Backfl??de",
	"130": "Permanent Fel L??g Batteriniv?? och Tempor??rt Fel Backfl??de",
	"144": "Tempor??rt Fel Is eller Frys Varning",
	"148": "Is eller Frys Varning och Tempor??rt Fel L??g Batteriniv??",
	"152": "Permanent Fel och Tempor??rt Fel Is eller Frys Varning",
	"156": "Permanent Fel L??g Batteriniv?? och Tempor??rt Fel Is eller Frys Varning",
	"176": "Tempor??rt Fel Spricka eller ??ppet R??r",
	"180": "Spricka eller ??ppet R??r och Tempor??rt Fel L??g Batteriniv??",
	"184": "Permanent Fel och Tempor??rt Fel Spricka eller ??ppet R??r",
	"188": "Permanent Fel L??g Batteriniv?? och Tempor??rt Fel Spricka eller ??ppet R??r",
}
