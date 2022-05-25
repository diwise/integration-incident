package application

import (
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/matryer/is"
	"github.com/rs/zerolog/log"
)

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", "0")
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceStateIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", "1")
	is.NoErr(err)
	incRep.assertNotCalled(is)

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", "1")
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentWhenUpdatedStateIsNoError(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", "1")
	is.NoErr(err)
	incRep.assertNotCalled(is)

	const stateNoError string = "0"
	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", stateNoError)
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceStateUpdatedSendsIncidentReportOnStateChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", "0")
	is.NoErr(err)

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", "4")
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "devId3 - Låg Batterinivå")
	is.Equal(incRep.incidents[0].Category, 17)
}

func TestThatDeviceUpdatedSendsIncidentReportEvenOnUnknownState(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", "0")
	is.NoErr(err)

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", "3")
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "devId4 - Okänt Fel: 3")
	is.Equal(incRep.incidents[0].Category, 17)
}

func TestThatDeviceValueUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.LifebuoyValueUpdated("urn:ngsi-ld:Device:elt-livboj-01", "off")
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceValueUpdatedDoesNotSendIncidentIfDeviceValueIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.LifebuoyValueUpdated("urn:ngsi-ld:Device:elt-livboj-01", "on")
	is.NoErr(err)
	incRep.assertNotCalled(is)

	err = app.LifebuoyValueUpdated("urn:ngsi-ld:Device:elt-livboj-01", "on")
	is.NoErr(err)
	incRep.assertNotCalled(is)
}

func TestThatDeviceValueUpdatedSendsIncidentReportOnValueChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.LifebuoyValueUpdated("urn:ngsi-ld:Device:se:servanet:lora:sn-elt-livboj-02", "on")
	is.NoErr(err)

	err = app.LifebuoyValueUpdated("urn:ngsi-ld:Device:se:servanet:lora:sn-elt-livboj-02", "off")
	is.NoErr(err)
	incRep.assertCalledOnce(is)
	is.Equal(incRep.incidents[0].Description, "Livboj kan ha flyttats eller utsatts för åverkan.")
	is.Equal(incRep.incidents[0].Category, 15)
}

func testSetup(t *testing.T) (*is.I, *incidentReporter, IntegrationIncident) {
	is := is.New(t)
	incRep := newIncidentReporterThatReturns(nil)
	app := NewApplication(log.Logger, incRep.f, "", "")

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

func (r *incidentReporter) f(incident models.Incident) error {
	r.callCount++
	r.incidents = append(r.incidents, incident)
	return r.returnValue
}
