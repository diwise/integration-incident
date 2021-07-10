package application

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diwise/integration-incident/incident"
	"github.com/diwise/integration-incident/infrastructure/logging"
)

func TestGetDeviceStatus(t *testing.T) {
	log := logging.NewLogger()

	server := setupMockService(http.StatusOK, livbojJson)

	nr := httptest.NewRecorder()

	incidentReporter, _ := incident.NewIncidentReporter(log, server.URL, "")

	GetDeviceStatusAndSendReportIfMissing(log, server.URL, incidentReporter)
	if nr.Code != http.StatusOK {
		t.Errorf("Request failed, status code not OK: %d", nr.Code)
	}

}

func setupMockService(responseCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(accessTokenResp))
		} else {
			w.Header().Add("Content-Type", "application/ld+json")
			w.WriteHeader(responseCode)
			w.Write([]byte(responseBody))
		}
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

const accessTokenResp string = `{"access_token":"ncjklhclabclksabclac",
"scope":"am_application_scope default",
"token_type":"Bearer",
"expires_in":3600}
`
