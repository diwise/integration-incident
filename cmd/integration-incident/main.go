package main

import (
	"os"
	"strings"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/presentation"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/rs/zerolog/log"
)

func main() {
	serviceName := "integration-incident"

	log := log.With().Str("service", strings.ToLower(serviceName)).Logger()
	log.Info().Msg("starting up ...")

	baseUrl := os.Getenv("DIWISE_BASE_URL")
	gatewayUrl := os.Getenv("GATEWAY_URL")
	authCode := os.Getenv("AUTH_CODE")
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	incidentReporter, err := incident.NewIncidentReporter(log, gatewayUrl, authCode)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create incident reporter")
	}

	app := application.NewApplication(log, incidentReporter, baseUrl, port)

	err = app.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start application")
	}

	err = presentation.CreateRouterAndStartServing(log, app, port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start router")
	}

}
