package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"
)

var deviceStatusCache map[string]string

func GetDeviceStatusAndSendReportIfMissing(log logging.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {

	devices, err := getDevicesFromContextBroker(log, baseUrl)
	if err != nil {
		log.Errorf("failed to get devices from context: %s", err.Error())
		return err
	}

	const lifebuoyCategory int = 15
	inc := models.Incident{}
	deviceStatusCache := make(map[string]string)

	for _, device := range devices {

		if strings.Contains(device.ID, "sn-elt-livboj-") {

			_, ok := deviceStatusCache[device.ID]

			if !ok {
				deviceStatusCache[device.ID] = device.Value.Value
				continue

			} else {

				if device.Value.Value == "off" && device.Value.Value != deviceStatusCache[device.ID] {

					inc.PersonId = device.ID

					if device.Location != nil {
						lon := device.Location.GetAsPoint().Coordinates[0]
						lat := device.Location.GetAsPoint().Coordinates[1]

						inc.MapCoordinates = fmt.Sprintf("%f,%f", lat, lon)
					}

					inc.Category = lifebuoyCategory
					inc.Description = "Livboj kan ha flyttats eller utsatts för åverkan."

					err = incidentReporter(inc)
					if err != nil {
						log.Errorf("could not post incident: %s", err.Error())
						return err
					}
				}
				deviceStatusCache[device.ID] = device.Value.Value
			}
		}
	}

	return nil
}

func getDevicesFromContextBroker(log logging.Logger, host string) ([]*fiware.Device, error) {
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
