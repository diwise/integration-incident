package main

import (
	"os"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/internal/pkg/application"
)

func main() {
	log := logging.NewLogger()

	baseUrl := os.Getenv("DIWISE_BASE_URL")
	gatewayUrl := os.Getenv("GATEWAY_URL")
	apiKey := os.Getenv("API_KEY")

	log.Infof("Polling for device status ...")

	application.GetDeviceStatus(log, baseUrl, gatewayUrl, apiKey)
}
