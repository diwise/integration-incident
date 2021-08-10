package main

import (
	"os"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/logging"
	"github.com/diwise/integration-incident/pkg/incident"
)

func main() {
	log := logging.NewLogger()

	baseUrl := os.Getenv("DIWISE_BASE_URL")
	gatewayUrl := os.Getenv("GATEWAY_URL")
	authCode := os.Getenv("AUTH_CODE")

	incidentReporter, err := incident.NewIncidentReporter(log, gatewayUrl, authCode)
	if err != nil {
		log.Fatalf("failed to create incident reporter: %s", err.Error())
	}

	err = application.Run(log, baseUrl, incidentReporter)
	if err != nil {
		log.Fatalf("failed to run application, %s", err.Error())
	}

}
