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

	deviceStatus := models.DeviceStatus{}
	incident := models.Incident{}

	for _, device := range devices {

		if strings.Contains(device.ID, "sn-elt-livboj-") {

			if device.Value.Value == "off" {

				if deviceStatus.DeviceId == device.ID && deviceStatus.Status == "on" {

					incident.DeviceId = device.ID

					if device.Location != nil {
						incident.Coordinates = device.Location.GetAsPoint().Coordinates
					}

					incident.Category = 5
					incident.Description = "Lifebuoy may have been moved or tampered with."

					fmt.Print(incident)

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

	fmt.Print(incidentBytes)

	resp, err := http.Post(gatewayUrl, "application/ld+json", bytes.NewBuffer(incidentBytes))
	if resp.StatusCode != http.StatusOK || err != nil {
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
