package presentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/presentation/api"
	"github.com/matryer/is"
)

func TestNotificationHandlerDoesNothingIfDeviceStateDoesNotExist(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(noDeviceState)))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerTriggersDeviceStateUpdatedIfDeviceStateExists(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBody("se:servanet:lora:msva:123", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 1)
}

func TestNotificationHandlerDoesNotTriggersDeviceStateUpdatedIfWrongDeviceID(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBody("notawatermeter", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerReturnsBadRequestIfEmptyRequestBody(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", nil)
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestNotificationHandlerReturnsBadRequestIfRequestBodyCannotBeUnmarshalledToNotification(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(badRequestJson)))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestNotificationHandlerHandlesUpdatedValueForLifeBuoys(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBodyWithValue("sn-elt-livboj-01", "on"))))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(len(app.LifebuoyValueUpdatedCalls()), 1)
}

func TestNotificationLD(t *testing.T) {
	is := is.New(t)
	n := api.Notification{}
	json.Unmarshal([]byte(lifebuoy_on), &n)

	if len(n.Data) != 0 {
		for _, r := range n.Data {

			t := r["type"].(string)
			f := r["status"].(string)

			is.Equal("Lifebuoy", t)
			is.Equal("on", f)
		}
	}
}

func createStatusBody(deviceId, state string) string {
	return fmt.Sprintf(withDeviceStateJsonFormat, deviceId, state)
}

func createStatusBodyWithValue(deviceId, value string) string {
	return fmt.Sprintf(withValueJsonFormat, deviceId, value)
}

func mockApp() *application.IntegrationIncidentMock {
	return &application.IntegrationIncidentMock{
		DeviceStateUpdatedFunc: func(deviceId, deviceState string) error {
			return nil
		},
		LifebuoyValueUpdatedFunc: func(deviceId, deviceValue string) error {
			return nil
		},
	}
}

const badRequestJson string = `{
	"id": "urn:ngsi-ld:Device:se:servanet:lora:msva:123",
	"type": "Device",
	"rssi": 0.1,
	"snr": 0.41
}`
const noDeviceState string = `{
	"subscriptionId": "36990e41ccd84af99d8b233eca81d1d3",
	"data": [{
		"id": "urn:ngsi-ld:Device:se:servanet:lora:msva:123",
		"type": "Device",
		"rssi": 0.1,
		"snr": 0.41
	}]
}`
const withDeviceStateJsonFormat string = `{
	"subscriptionId": "36990e41ccd84af99d8b233eca81d1d3",
	"data": [
		{
			"id": "urn:ngsi-ld:Device:%s",
			"type": "Device",
			"rssi": 0.1,	
			"snr": 0.41,
			"deviceState": "%s"
		}
	]
}`
const withValueJsonFormat string = `{
	"subscriptionId": "36990e41ccd84af99d8b233eca81d1d3",
	"data": [
		{
			"id": "urn:ngsi-ld:Lifebuoy:%s",
			"type": "Lifebuoy",
			"rssi": 0.1,
			"snr":  0.41,		
			"status": "%s"		
		}
	]
}`

const lifebuoy_on string = `
{
	"id": "urn:ngsi-ld:Notification:628cce184ed0912f6cad226e",
	"type": "Notification",
	"subscriptionId": "urn:ngsi-ld:Subscription:628ccdd44ed0912f6cad226d",
	"notifiedAt": "2022-05-24T12:22:48.505Z",
	"data": [
		{
			"id": "urn:ngsi-ld:Lifebuoy:deviceID-001",
			"type": "Lifebuoy",
			"status": "on"	  
		}
	]
}
`
