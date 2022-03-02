package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"

	"github.com/rs/zerolog"
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
	stateChanged := a.checkDeviceExistsAndPreviousDeviceState(deviceId, deviceState)

	if !stateChanged {
		return nil
	}

	err := a.createAndSendIncident(deviceId, deviceState, a.incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to create and send incident: %s", err.Error())
	}

	a.updateDeviceState(deviceId, deviceId)

	return nil
}

func (a *app) updateDeviceState(deviceId, deviceState string) {
	a.previousStates[deviceId] = deviceState
}
