package application

import (
	"time"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/infrastructure/repositories/models"
)

func Run(log logging.Logger, baseUrl string, incidentReporter func(models.Incident) error) {
	log.Infof("Polling for device status ...")

	for {
		GetDeviceStatusAndSendReportIfMissing(log, baseUrl, incidentReporter)
		time.Sleep(5 * time.Second)
	}
}
