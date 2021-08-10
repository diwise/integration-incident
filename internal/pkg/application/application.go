package application

import (
	"fmt"
	"time"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/infrastructure/repositories/models"
)

func Run(log logging.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {
	err := GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
	if err != nil {
		return fmt.Errorf("failed to start polling for devices: %s", err.Error())
	}

	for {
		log.Infof("Polling for device status ...")
		GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
		time.Sleep(5 * time.Second)
	}
}
