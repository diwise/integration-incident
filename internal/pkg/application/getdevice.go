package application

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"
	"github.com/rs/zerolog"
)

func getDeviceFromContextBroker(log zerolog.Logger, host, deviceId string) (*fiware.Device, error) {
	log.Info().Msgf("requesting device details for %s from %s", deviceId, host)

	response, err := http.Get(fmt.Sprintf("%s/ngsi-ld/v1/entities/%s", host, deviceId))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %d", response.StatusCode)
	}

	defer response.Body.Close()

	device := &fiware.Device{}

	err = json.NewDecoder(response.Body).Decode(&device)

	return device, err
}
