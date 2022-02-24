package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"
	"github.com/rs/zerolog"
)

var deviceStatusCache map[string]string = make(map[string]string)

func GetDeviceStatusAndSendReportIfMissing(log zerolog.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {

	devices, err := getDevicesFromContextBroker(log, baseUrl)
	if err != nil {
		log.Err(err).Msgf("failed to get devices from context: %s", err.Error())
		return err
	}

	const lifebuoyCategory int = 15
	inc := models.Incident{}

	for _, device := range devices {

		if strings.Contains(device.ID, "sn-elt-livboj-") {

			_, ok := deviceStatusCache[device.ID]

			if !ok {
				deviceStatusCache[device.ID] = device.Value.Value
				continue
			} else {

				storedValue := deviceStatusCache[device.ID]

				if device.Value.Value == "off" && device.Value.Value != storedValue {

					log.Info().Msgf("state changed to \"off\" for device %s", device.ID)

					inc.PersonId = "diwise"

					if device.Location != nil {
						lon := device.Location.GetAsPoint().Longitude()
						lat := device.Location.GetAsPoint().Latitude()

						inc.MapCoordinates = fmt.Sprintf("%f,%f", lat, lon)
					}

					inc.Category = lifebuoyCategory
					inc.Description = "Livboj kan ha flyttats eller utsatts för åverkan."

					err = incidentReporter(inc)
					if err != nil {
						log.Err(err).Msgf("could not post incident: %s", err.Error())
						return err
					}
				}
				deviceStatusCache[device.ID] = device.Value.Value
			}
		}
	}

	return nil
}

func getDevicesFromContextBroker(log zerolog.Logger, host string) ([]*fiware.Device, error) {
	response, err := http.Get(fmt.Sprintf("%s/ngsi-ld/v1/entities?type=Device", host))
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %d", response.StatusCode)
	}

	defer response.Body.Close()

	devices := []*fiware.Device{}

	err = json.NewDecoder(response.Body).Decode(&devices)

	return devices, err
}
