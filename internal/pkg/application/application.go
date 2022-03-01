package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/integration-incident/internal/pkg/presentation"
	"github.com/rs/zerolog"
)

type Application interface {
	Start() error
	RunPoll(log zerolog.Logger, baseUrl string, incidentReporter func(models.Incident) error) error
}

type newIntegrationIncident struct {
	log              zerolog.Logger
	incidentReporter func(models.Incident) error
	baseUrl          string
	port             string
}

func NewApplication(log zerolog.Logger, incidentReporter func(models.Incident) error, baseUrl, port string) newIntegrationIncident {
	return newIntegrationIncidentApp(log, incidentReporter, baseUrl, port)
}

func newIntegrationIncidentApp(log zerolog.Logger, incidentReporter func(models.Incident) error, baseUrl, port string) newIntegrationIncident {
	app := newIntegrationIncident{
		log:              log,
		incidentReporter: incidentReporter,
		baseUrl:          baseUrl,
		port:             port,
	}

	return app
}

func (a *newIntegrationIncident) Start() error {

	go a.RunPoll(a.log, a.baseUrl, a.incidentReporter)

	err := presentation.CreateRouterAndStartServing(a.log, a.incidentReporter, a.port)
	if err != nil {
		return err
	}

	return nil
}

func (a *newIntegrationIncident) RunPoll(log zerolog.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {
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
