package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/presentation"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/diwise/service-chassis/pkg/infrastructure/buildinfo"
	"github.com/diwise/service-chassis/pkg/infrastructure/env"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
)

const serviceName string = "integration-incident"

func main() {

	serviceVersion := buildinfo.SourceVersion()
	ctx, logger, cleanup := o11y.Init(context.Background(), serviceName, serviceVersion)
	defer cleanup()

	baseUrl := os.Getenv("DIWISE_BASE_URL")

	gatewayUrl := env.GetVariableOrDie(ctx, "GATEWAY_URL", "valid gateway URL")
	authCode := env.GetVariableOrDie(ctx, "AUTH_CODE", "valid auth code")
	port := env.GetVariableOrDefault(ctx, "SERVICE_PORT", "8080")
	tenant := env.GetVariableOrDefault(ctx, "DIWISE_TENANT", "default")

	incidentReporter, err := incident.NewIncidentReporter(ctx, gatewayUrl, authCode)
	if err != nil {
		fatal(ctx, "failed to create incident reporter", err)
	}

	entityLocator, err := services.NewEntityLocator(baseUrl, tenant)
	if err != nil {
		fatal(ctx, "failed to create entity locator", err)
	}

	app := application.NewApplication(ctx, incidentReporter, entityLocator)

	mux, err := presentation.CreateRouter(ctx, app)
	if err != nil {
		fatal(ctx, "failed to start router", err)
	}

	webServer := &http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start request router", "err", err.Error())
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigChan

	logger.Debug("received signal", "signal", s)

	err = webServer.Shutdown(ctx)
	if err != nil {
		logger.Error("failed to shutdown web server", "err", err.Error())
	}

	logger.Info("shutting down")
}

func fatal(ctx context.Context, msg string, err error) {
	logging.GetFromContext(ctx).Error(msg, "err", err.Error())
	os.Exit(1)
}
