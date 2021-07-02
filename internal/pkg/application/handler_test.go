package application

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-incident/infrastructure/logging"
	"github.com/diwise/integration-incident/infrastructure/repositories/models"
)

func TestGetDeviceStatus(t *testing.T) {
	log := logging.NewLogger()

	server := setupMockService(http.StatusOK, livbojJson)

	nr := httptest.NewRecorder()

	GetDeviceStatus(log, server.URL, server.URL, "")
	if nr.Code != http.StatusOK {
		t.Errorf("Request failed, status code not OK: %d", nr.Code)
	}

}

func TestPostIncident(t *testing.T) {
	log := logging.NewLogger()

	server := setupMockService(http.StatusOK, "")

	incident := models.Incident{
		PersonId:       "deviceID",
		Description:    "description",
		Category:       5,
		MapCoordinates: "62.0,17.0",
	}

	PostIncident(log, incident, server.URL, "")
}

func setupMockService(responseCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/ld+json")
		w.WriteHeader(responseCode)
		w.Write([]byte(responseBody))
	}))
}

const livbojJson string = `[
	{
		"@context": [
			"https://schema.lab.fiware.org/ld/context",
			"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
		],
		"id": "urn:ngsi-ld:Device:se:servanet:lora:sn-elt-livboj-01",
		"type": "Device",
		"value": {
			"type": "Property",
			"value": "on"
		}
	},
	{
		"@context": [
			"https://schema.lab.fiware.org/ld/context",
			"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
		],
		"id": "urn:ngsi-ld:Device:se:servanet:lora:sn-elt-livboj-01",
		"type": "Device",
		"value": {
			"type": "Property",
			"value": "off"
		}
	}
]`

/*
	{
		"@context": [
			"https://schema.lab.fiware.org/ld/context",
			"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
		],
		"id": "urn:ngsi-ld:Device:se:servanet:lora:sn-elt-livboj-02”,
		"refDeviceModel": {
			"object": "urn:ngsi-ld:DeviceModel:se:elsys:elt-lite:temp",
			"type": "Relationship"
		},
		"type": "Device",
		"value": {
			"type": "Property",
			"value": “off”
		}
	}
*/
