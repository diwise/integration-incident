package application

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/rs/zerolog"
)

func getLifebuoyFromContextBroker(log zerolog.Logger, host, deviceId string) (*diwise.Lifebuoy, error) {
	log.Info().Msgf("requesting lifebuoy details for %s from %s", deviceId, host)

	response, err := http.Get(fmt.Sprintf("%s/ngsi-ld/v1/entities/%s", host, deviceId))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %d != %d", response.StatusCode, http.StatusOK)
	}

	lifebuoy := &diwise.Lifebuoy{}

	err = json.NewDecoder(response.Body).Decode(&lifebuoy)

	return lifebuoy, err
}
