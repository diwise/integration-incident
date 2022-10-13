package presentation

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestNotificationHandlerDoesNothingIfDeviceStateDoesNotExist(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(noDeviceState)))
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerTriggersDeviceStateUpdatedIfDeviceStateExists(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBody("se:servanet:lora:msva:123", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 1)
}

func TestNotificationHandlerDoesNotTriggersDeviceStateUpdatedIfWrongDeviceID(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBody("notawatermeter", "104"))))
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	is.Equal(len(app.DeviceStateUpdatedCalls()), 0)
}

func TestNotificationHandlerReturnsBadRequestIfEmptyRequestBody(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", nil)
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestNotificationHandlerReturnsBadRequestIfRequestBodyCannotBeUnmarshalledToNotification(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(badRequestJson)))
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusBadRequest)
}

func TestNotificationHandlerHandlesUpdatedValueForLifeBuoys(t *testing.T) {
	is := is.New(t)
	app := mockApp()

	r := httptest.NewRequest("POST", "/api/notify", bytes.NewBuffer([]byte(createStatusBodyWithValue("sn-elt-livboj-01", "on"))))
	w := httptest.NewRecorder()

	notificationHandler(zerolog.Logger{}, app).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)
	is.Equal(len(app.LifebuoyValueUpdatedCalls()), 1)
}

func createStatusBody(deviceId, state string) string {
	return fmt.Sprintf(withDeviceStateJsonFormat, deviceId, state)
}

func createStatusBodyWithValue(deviceId, value string) string {
	return fmt.Sprintf(withValueJsonFormat, deviceId, value)
}

func mockApp() *application.IntegrationIncidentMock {
	return &application.IntegrationIncidentMock{
		DeviceStateUpdatedFunc: func(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error {
			return nil
		},
		LifebuoyValueUpdatedFunc: func(ctx context.Context, deviceId, deviceValue string) error {
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
			"deviceState": {
				"type": "Property",
				"value": "%s"
			}
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
			"status": {
				"type": "Property",
				"value": "%s"
			}
		}
	]
}`
