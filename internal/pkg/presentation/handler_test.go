package presentation

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/matryer/is"
)

func TestNotificationHandlerDoesNothingIfDeviceStateDoesNotExist(t *testing.T) {
	is := is.New(t)

	server := setupMockService(http.StatusOK, "")

	r := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(noDeviceState)))
	w := httptest.NewRecorder()

	reporter := newIncidentReporterThatReturns(nil)

	notificationHandler(reporter.f).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)
}

func TestNotificationHandlerAddsNewDeviceStateIfNoPreviousStateExists(t *testing.T) {
	is := is.New(t)

	server := setupMockService(http.StatusOK, "")

	r := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(createStatusBody("123", "104"))))
	w := httptest.NewRecorder()

	reporter := newIncidentReporterThatReturns(nil)

	notificationHandler(reporter.f).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	//this test doesn't actually check that a new device state has been stored yet
}

func TestNotificationHandlerSendsNewIncidentReportIfDeviceStateHasChanged(t *testing.T) {
	is := is.New(t)

	server := setupMockService(http.StatusOK, "")

	r := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(createStatusBody("123", "104"))))
	w := httptest.NewRecorder()

	reporter := newIncidentReporterThatReturns(nil)

	notificationHandler(reporter.f).ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)

	r2 := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(createStatusBody("123", "205"))))

	notificationHandler(reporter.f).ServeHTTP(w, r2)
}

func setupMockService(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/ld+json")
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

const noDeviceState string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:05393925","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41}}]}`
const withDeviceStateJsonFormat string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:%s","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41},"deviceState":{"type":"Property","value":"%s"}}]}`

func createStatusBody(deviceId, state string) string {

	return fmt.Sprintf(withDeviceStateJsonFormat, deviceId, state)
}

func newIncidentReporterThatReturns(err error) *incidentReporter {
	return &incidentReporter{returnValue: err}
}

type incidentReporter struct {
	callCount   int32
	returnValue error
}

func (r *incidentReporter) assertCallCount(t *testing.T, expected int32) {
	if r.callCount != expected {
		if expected == 0 {
			t.Errorf("Incident reporter should not have been called, but was called %d times!", r.callCount)
		} else if expected == 1 {
			t.Errorf("Incident reporter should have been called once, but was called %d times!", r.callCount)
		} else {
			t.Errorf("Incident reporter should have been called %d times, but was called %d times!", expected, r.callCount)
		}
	}
}

func (r *incidentReporter) assertCalledOnce(t *testing.T) {
	r.assertCallCount(t, 1)
}

func (r *incidentReporter) assertNotCalled(t *testing.T) {
	r.assertCallCount(t, 0)
}

func (r *incidentReporter) f(models.Incident) error {
	r.callCount++
	return r.returnValue
}
