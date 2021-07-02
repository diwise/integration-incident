package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/infrastructure/repositories/models"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"
)

func GetDeviceStatus(log logging.Logger, baseUrl, gatewayUrl, apiKey string) error {

	devices, err := getDevicesFromContextBroker(baseUrl)
	if err != nil {
		log.Errorf("failed to get devices from context: %s", err)
		return err
	}

	const lifebuoyCategory int = 15
	deviceStatus := models.DeviceStatus{}
	incident := models.Incident{}

	for _, device := range devices {

		if strings.Contains(device.ID, "sn-elt-livboj-") {

			if device.Value.Value == "off" {

				if deviceStatus.DeviceId == device.ID && deviceStatus.Status == "on" {

					incident.PersonId = device.ID

					if device.Location != nil {
						lon := device.Location.GetAsPoint().Coordinates[0]
						lat := device.Location.GetAsPoint().Coordinates[1]

						incident.MapCoordinates = fmt.Sprintf("%f,%f", lat, lon)
					}

					incident.Category = lifebuoyCategory
					incident.Description = "Livboj kan ha flyttats eller utsatts för åverkan."

					PostIncident(log, incident, gatewayUrl, apiKey)
				}
			}

			deviceStatus.DeviceId = device.ID
			deviceStatus.Status = device.Value.Value
		}
	}

	return nil
}

func PostIncident(log logging.Logger, incident models.Incident, gatewayUrl, apiKey string) error {

	incidentBytes, err := json.Marshal(incident)
	if err != nil {
		log.Errorf("could not marshal incident message into json")
	}

	fmt.Println(string(incidentBytes))
	log.Infof("posting incident to: %s", gatewayUrl)

	resp, err := http.Post(gatewayUrl, "application/ld+json", bytes.NewBuffer(incidentBytes))
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Errorf("failed to post incident message: %s", err)
		return err
	}

	return nil
}

func getDevicesFromContextBroker(host string) ([]*fiware.Device, error) {
	response, err := http.Get(fmt.Sprintf("%s/ngsi-ld/v1/entities?type=Device", host))
	if response.StatusCode != http.StatusOK {
		return nil, err
	}

	defer response.Body.Close()

	devices := []*fiware.Device{}

	err = json.NewDecoder(response.Body).Decode(&devices)

	return devices, err
}
