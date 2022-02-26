package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/rs/zerolog"
)

func RunPoll(log zerolog.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {
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
