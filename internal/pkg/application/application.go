package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/logging"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
)

func Run(log logging.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {
	log.Infof("Polling for device status ...")

	err := GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to start polling for devices, %s", err.Error())
	}

	for {
		GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
		time.Sleep(5 * time.Second)
	}
}
