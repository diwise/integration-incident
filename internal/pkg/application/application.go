package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"

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
}

func NewApplication(log zerolog.Logger, incidentReporter func(models.Incident) error, baseUrl, port string) IntegrationIncident {
	newApp := &app{
		log:              log,
		incidentReporter: incidentReporter,
		baseUrl:          baseUrl,
		port:             port,
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
	stateChanged := checkIfDeviceExistsAndPreviousDeviceState(deviceId, deviceState)

	if !stateChanged {
		log.Info().Msg("device either does not exist, or state has not changed...")
		return nil
	}

	err := createAndSendIncident(deviceId, deviceState, a.incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to create and send incident: %s", err.Error())
	}

	return nil
}
