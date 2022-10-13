package main

import (
	"context"
	"os"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/presentation"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
)

const serviceName string = "integration-incident"

func main() {

	serviceVersion := buildinfo.SourceVersion()
	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion)
	defer cleanup()

	baseUrl := os.Getenv("DIWISE_BASE_URL")

	gatewayUrl := env.GetVariableOrDie(logger, "GATEWAY_URL", "valid gateway URL")
	authCode := env.GetVariableOrDie(logger, "AUTH_CODE", "valid auth code")
	port := env.GetVariableOrDefault(logger, "SERVICE_PORT", "8080")

	incidentReporter, err := incident.NewIncidentReporter(ctx, gatewayUrl, authCode)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create incident reporter")
	}

	app := application.NewApplication(ctx, incidentReporter, baseUrl, port)

	err = presentation.CreateRouterAndStartServing(ctx, app, port)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to start router")
	}

}
