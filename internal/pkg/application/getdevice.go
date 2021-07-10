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

func GetDeviceStatusAndSendReportIfMissing(log logging.Logger, baseUrl string, incidentReporter func(models.Incident) error) error {

	devices, err := getDevicesFromContextBroker(baseUrl)
	if err != nil {
		log.Errorf("failed to get devices from context: %s", err.Error())
		return err
	}

	const lifebuoyCategory int = 15
	deviceStatus := models.DeviceStatus{}
	inc := models.Incident{}

	for _, device := range devices {

		if strings.Contains(device.ID, "sn-elt-livboj-") {

			if device.Value.Value == "off" {

				if deviceStatus.DeviceId == device.ID && deviceStatus.Status == "on" {

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
					}
				}
			}

			deviceStatus.DeviceId = device.ID
			deviceStatus.Status = device.Value.Value
		}
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
