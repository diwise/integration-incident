package application

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/matryer/is"
)

func status(deviceID string, code int, timestamp time.Time, messages ...string) models.StatusMessage {
	bat := float64(100)
	c := strconv.Itoa(code)

	return models.StatusMessage{
		DeviceID:     deviceID,
		Timestamp:    timestamp,
		BatteryLevel: &bat,
		Messages:     messages,
		Code:         &c,
	}
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfStatusIsNoErrOrPayloadErr(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 0, time.Now().UTC(), "No error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 100, time.Now().UTC(), "Payload error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfStatusIsFreezeAndOutOfBounds(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 1, time.Date(2024, 5, 31, 12, 12, 12, 0, time.UTC), "Freeze"))
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedSendIncidentIfStatusIsFreezeAndWithinBounds(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 1, time.Date(2024, 2, 28, 12, 12, 12, 0, time.UTC), "Freeze"))
	is.NoErr(err)
	incRep.assertCalledOnce(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", 0, time.Now().UTC(), "No error"))
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceStateIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1, time.Now().UTC()))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1, time.Now().UTC()))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentWhenUpdatedStateIsNoError(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 1, time.Now().UTC()))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", 0, time.Now().UTC(), "No error"))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentReportOnStateChangedButWrongErrorType(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", 0, time.Now().UTC()))
	is.NoErr(err)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", 4, time.Now().UTC(), "Power low"))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)
}

func TestThatDeviceUpdatedDoesNotSendIncidentReportOnUnknownState(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", 0, time.Now().UTC()))
	is.NoErr(err)

	err = app.DeviceStateUpdated(context.Background(), "urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", status("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", 3, time.Now().UTC(), "Unknown"))
	is.NoErr(err)
	incRep.assertCallCount(is, 0)
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
