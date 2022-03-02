package presentation

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/matryer/is"
)

func TestNotificationHandlerDoesNothingIfDeviceStateDoesNotExist(t *testing.T) {
	is := is.New(t)

	app := mockApp()

	r := httptest.NewRequest("POST", "/notification", bytes.NewBuffer([]byte(noDeviceState)))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerTriggersDeviceStateUpdatedIfDeviceStateExists(t *testing.T) {
	is := is.New(t)

	app := mockApp()

	r := httptest.NewRequest("POST", "/notification", bytes.NewBuffer([]byte(createStatusBody("se:servanet:lora:msva:123", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 1)
}

func TestNotificationHandlerDoesNotTriggersDeviceStateUpdatedIfWrongDeviceID(t *testing.T) {
	is := is.New(t)

	app := mockApp()

	r := httptest.NewRequest("POST", "/notification", bytes.NewBuffer([]byte(createStatusBody("notawatermeter", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerReturnsBadRequestIfEmptyRequestBody(t *testing.T) {
	is := is.New(t)

	app := mockApp()

	r := httptest.NewRequest("POST", "/notification", nil)
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestNotificationHandlerReturnsBadRequestIfRequestBodyCannotBeUnmarshalledToNotification(t *testing.T) {
	is := is.New(t)

	app := mockApp()

	r := httptest.NewRequest("POST", "/notification", bytes.NewBuffer([]byte(badRequestJson)))
	w := httptest.NewRecorder()

	notificationHandler(app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

const badRequestJson string = `{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:123","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41}}`
const noDeviceState string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:123","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41}}]}`
const withDeviceStateJsonFormat string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:%s","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41},"deviceState":{"type":"Property","value":"%s"}}]}`

func createStatusBody(deviceId, state string) string {
	return fmt.Sprintf(withDeviceStateJsonFormat, deviceId, state)
}

func mockApp() *application.IntegrationIncidentMock {
	return &application.IntegrationIncidentMock{
		DeviceStateUpdatedFunc: func(deviceId, deviceState string) error {
			return nil
		},
	}
}
