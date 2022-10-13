package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"

	"github.com/rs/zerolog"
)

//go:generate moq -rm -out application_mock.go . IntegrationIncident

type IntegrationIncident interface {
	DeviceStateUpdated(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error
	LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error
}

type app struct {
	log              zerolog.Logger
	incidentReporter func(context.Context, models.Incident) error
	baseUrl          string
	port             string
	previousStates   map[string]string
	previousValues   map[string]string
}

func NewApplication(ctx context.Context, incidentReporter func(context.Context, models.Incident) error, baseUrl, port string) IntegrationIncident {

	newApp := &app{
		log:              logging.GetFromContext(ctx),
		incidentReporter: incidentReporter,
		baseUrl:          baseUrl,
		port:             port,
		previousStates:   make(map[string]string),
		previousValues:   make(map[string]string),
	}

	return newApp
}

func (a *app) DeviceStateUpdated(ctx context.Context, deviceId string, sm models.StatusMessage) error {

	if !strings.Contains(deviceId, "se:servanet:lora:msva:") {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	shortId := deviceId[strings.LastIndex(deviceId, ":")+1:]

	deviceState := strconv.Itoa(sm.Status)

	if !a.checkIfDeviceStateHasChanged(shortId, deviceState) {
		return nil
	}

	const stateNoError string = "0"

	if deviceState != stateNoError {
		const watermeterCategory int = 17
		incident := models.NewIncident(watermeterCategory, translateJoin(shortId, sm)).AtLocation(62.388178, 17.315090)

		err := a.incidentReporter(ctx, *incident)
		if err != nil {
			a.log.Err(err).Msg("could not post incident")
			return err
		}
	}

	a.updateDeviceState(shortId, deviceState)

	return nil
}

func (a *app) LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error {
	if !strings.HasPrefix(deviceId, diwise.LifebuoyIDPrefix) {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	shortId := strings.TrimPrefix(deviceId, diwise.LifebuoyIDPrefix)
	valueChanged := a.checkIfDeviceValueHasChanged(shortId, deviceValue)

	if !valueChanged {
		return nil
	}

	if deviceValue == "off" {
		a.log.Info().Msgf("state changed to \"off\" on device: %s", shortId)

		const lifebuoyCategory int = 15
		incident := models.NewIncident(lifebuoyCategory, "Livboj kan ha flyttats eller utsatts för åverkan.")

		lifebuoy, err := getLifebuoyFromContextBroker(a.log, a.baseUrl, deviceId)

		if err == nil {
			point := lifebuoy.Location.GetAsPoint()
			incident = incident.AtLocation(point.Latitude(), point.Longitude())
		}

		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			a.log.Err(err).Msg("could not post incident")
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
		a.log.Info().Msgf("device %s does not exist, saving state...", deviceId)
		a.previousStates[deviceId] = state
		return false
	}

	if storedState != state {
		a.log.Info().Msgf("device %s state has changed from %s to %s", deviceId, storedState, state)
		return true
	}

	return false
}

func (a *app) checkIfDeviceValueHasChanged(deviceId, value string) bool {
	storedValue, exists := a.previousValues[deviceId]

	if !exists {
		a.log.Info().Msgf("device %s does not exist, saving value...", deviceId)
		a.previousValues[deviceId] = value
		return false
	}

	if storedValue != value {
		a.log.Info().Msgf("device %s value has changed to %s", deviceId, value)
		return true
	}

	return false
}

func translateJoin(deviceID string, sm models.StatusMessage) string {
	return fmt.Sprintf("%s - %s", deviceID, Join(sm.Messages, " ", translate))
}

func Join(elems []string, sep string, mod func(string) string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return mod(elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(mod(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(mod(s))
	}
	return b.String()
}

func translate(s string) string {
	switch s {
	case "No error":
		return "Inga fel"
	case "Power low":
		return "Låg batterinivå"
	case "Permanent error":
		return "Permanent fel"
	case "Temporary error":
		return "Temporärt fel"
	case "Empty spool":
		return "Tomt rör"
	case "Leak":
		return "Läckage"
	case "Burst":
		return "Spricka"
	case "Backflow":
		return "Backflöde"
	case "Freeze":
		return "Is eller Frys Varning"
	case "Unknown":
		return "Okänt fel"
	}
	return s
}
