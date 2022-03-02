package application

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/rs/zerolog"
)

func TestThatNoReportIsSentOnFirstUpdate(t *testing.T) {
	server := setupMockService([]response{
		{http.StatusOK, livbojJsonOneMissing},
	})
	reporter := newIncidentReporterThatReturns(nil)

	err := GetDeviceStatusAndSendReportIfMissing(zerolog.Logger{}, server.URL, reporter.f)

	if err != nil {
		t.Errorf("GetDeviceStatusAndSendReportIfMissing failed unexpectedly: %s", err.Error())
	} else {
		reporter.assertNotCalled(t)
	}
}

func TestThatAReportIsSentWhenOneIsMissing(t *testing.T) {
	server := setupMockService([]response{
		{http.StatusOK, livbojJsonAllPresent},
		{http.StatusOK, livbojJsonOneMissing},
	})
	reporter := newIncidentReporterThatReturns(nil)
	log := zerolog.Logger{}

	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	err := GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)

	if err != nil {
		t.Errorf("GetDeviceStatusAndSendReportIfMissing failed unexpectedly: %s", err.Error())
	} else {
		reporter.assertCalledOnce(t)
	}
}

func TestThatOffStateIsRememberedAndOnlyOneReportIsSent(t *testing.T) {
	server := setupMockService([]response{
		{http.StatusOK, livbojJsonAllPresent},
		{http.StatusOK, livbojJsonOneMissing},
		{http.StatusOK, livbojJsonOneMissing},
	})
	reporter := newIncidentReporterThatReturns(nil)
	log := zerolog.Logger{}

	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	err := GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)

	if err != nil {
		t.Errorf("GetDeviceStatusAndSendReportIfMissing failed unexpectedly: %s", err.Error())
	} else {
		reporter.assertCalledOnce(t)
	}
}

func TestThatANewReportIsSentAfterStateReset(t *testing.T) {
	server := setupMockService([]response{
		{http.StatusOK, livbojJsonAllPresent},
		{http.StatusOK, livbojJsonOneMissing},
		{http.StatusOK, livbojJsonAllPresent},
		{http.StatusOK, livbojJsonOneMissing},
	})
	reporter := newIncidentReporterThatReturns(nil)
	log := zerolog.Logger{}

	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)
	err := GetDeviceStatusAndSendReportIfMissing(log, server.URL, reporter.f)

	if err != nil {
		t.Errorf("GetDeviceStatusAndSendReportIfMissing failed unexpectedly: %s", err.Error())
	} else {
		reporter.assertCallCount(t, 2)
	}
}

type response struct {
	code int
	body string
}

func setupMockService(resp []response) *httptest.Server {
	responseIndex := 0

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with the current response ...
		w.Header().Add("Content-Type", "application/ld+json")
		w.WriteHeader(resp[responseIndex].code)
		w.Write([]byte(resp[responseIndex].body))

		// ... and move forward to the next response (start over if the call count exceeds the response count)
		responseIndex = (responseIndex + 1) % len(resp)
	}))
}

func newIncidentReporterThatReturns(err error) *incidentReporter {
	return &incidentReporter{returnValue: err}
}

type incidentReporter struct {
	callCount   int32
	returnValue error
	incidents   []models.Incident
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

func (r *incidentReporter) f(incident models.Incident) error {
	r.callCount++
	r.incidents = append(r.incidents, incident)
	return r.returnValue
}

var livbojJsonAllPresent string = createStatusBody("boj1", "on", "boj2", "on")
var livbojJsonOneMissing string = createStatusBody("boj1", "on", "boj2", "off")

func createStatusBody(args ...string) string {
	var deviceStatuses []string

	for i := 0; i < len(args); i += 2 {
		deviceStatuses = append(deviceStatuses, fmt.Sprintf(livbojJsonFormat, args[i], args[i+1]))
	}

	return fmt.Sprintf("[%s]", strings.Join(deviceStatuses, ","))
}

const livbojJsonFormat string = `{
	"@context": [
		"https://schema.lab.fiware.org/ld/context",
		"https://uri.etsi.org/ngsi-ld/v1/ngsi-ld-core-context.jsonld"
	],
	"id": "urn:ngsi-ld:Device:io:diwise:sn-elt-livboj-%s",
	"type": "Device",
	"value": {
		"type": "Property",
		"value": "%s"
	}
}`
