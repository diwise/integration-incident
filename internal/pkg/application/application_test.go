package application

import (
	"context"
	"testing"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/matryer/is"
)

func status(deviceID string, code int, messages ...string) models.StatusMessage {
	return models.StatusMessage{
		DeviceID:     deviceID,
		Timestamp:    time.Now().Format(time.RFC3339),
		BatteryLevel: 100,
		Messages:     messages,
		Code:       code,
	}
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfStatusIsNoErrOrPayloadErr(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 0, "No error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 100, "Payload error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)

}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 0, "No error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceStateIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1))
	is.NoErr(err)
	incRep.assertCallCount(is, 1)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1))
	is.NoErr(err)
	incRep.assertCallCount(is, 1)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentWhenUpdatedStateIsNoError(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1))
	is.NoErr(err)
	incRep.assertCallCount(is, 1)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 0, "No error"))
	is.NoErr(err)
	incRep.assertCallCount(is, 1)
}

func TestThatDeviceStateUpdatedSendsIncidentReportOnStateChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", 0))
	is.NoErr(err)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", 4, "Power low"))
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "devId3 - Låg batterinivå")
	is.Equal(incRep.incidents[0].Category, 17)
}

func TestThatDeviceUpdatedSendsIncidentReportEvenOnUnknownState(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", 0))
	is.NoErr(err)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", 3, "Unknown"))
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "devId4 - Okänt fel")
	is.Equal(incRep.incidents[0].Category, 17)
}

func TestThatDeviceValueUpdatedDoesNotSendIncidentIfDeviceValueIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.LifebuoyValueUpdated(context.Background(), "urn:ngsi-ld:Lifebuoy:elt-livboj-01", "on")
	is.NoErr(err)
	incRep.assertNotCalled(is)

	err = app.LifebuoyValueUpdated(context.Background(), "urn:ngsi-ld:Lifebuoy:elt-livboj-01", "on")
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceValueUpdatedSendsIncidentReportOnValueChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.LifebuoyValueUpdated(context.Background(), "urn:ngsi-ld:Lifebuoy:se:servanet:lora:sn-elt-livboj-02", "on")
	is.NoErr(err)

	err = app.LifebuoyValueUpdated(context.Background(), "urn:ngsi-ld:Lifebuoy:se:servanet:lora:sn-elt-livboj-02", "off")
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "Livboj kan ha flyttats eller utsatts för åverkan.")
	is.Equal(incRep.incidents[0].Category, 15)
}

func testSetup(t *testing.T) (*is.I, *incidentReporter, IntegrationIncident) {
	is := is.New(t)
	incRep := newIncidentReporterThatReturns(nil)
	locator := &services.EntityLocatorMock{
		LocateFunc: func(ctx context.Context, entityType, entityID string) (float64, float64, error) {
			return 62.388178, 17.315090, nil
		},
	}

	app := NewApplication(context.Background(), incRep.f, locator)

	return is, incRep, app
}

func newIncidentReporterThatReturns(err error) *incidentReporter {
	return &incidentReporter{returnValue: err}
}

type incidentReporter struct {
	callCount   int32
	returnValue error
	incidents   []models.Incident
}

func (r *incidentReporter) assertCallCount(is *is.I, expected int32) {
	is.Equal(r.callCount, expected) // missmatching call count
}

func (r *incidentReporter) assertCalledOnce(is *is.I) {
	r.assertCallCount(is, 1)
}

func (r *incidentReporter) assertNotCalled(is *is.I) {
	r.assertCallCount(is, 0)
}

func (r *incidentReporter) f(ctx context.Context, incident models.Incident) error {
	r.callCount++
	r.incidents = append(r.incidents, incident)
	return r.returnValue
}
