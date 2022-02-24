package main

import (
	"os"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.Logger{}

	baseUrl := os.Getenv("DIWISE_BASE_URL")
	gatewayUrl := os.Getenv("GATEWAY_URL")
	authCode := os.Getenv("AUTH_CODE")

	incidentReporter, err := incident.NewIncidentReporter(log, gatewayUrl, authCode)
	if err != nil {
		log.Fatal().Msgf("failed to create incident reporter: %s", err.Error())
	}

	err = application.Run(log, baseUrl, incidentReporter)
	if err != nil {
		log.Fatal().Msgf("failed to create incident reporter: %s", err.Error())
	}

}
