package application

import (
	"testing"

	"github.com/matryer/is"
	"github.com/rs/zerolog/log"
)

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId1", "0")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceStateIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", "0")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId2", "0")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))
}

func TestThatDeviceStateUpdatedSendsIncidentReportOnStateChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", "0")
	is.NoErr(err)

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId3", "4")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(1))
	is.Equal(incRep.incidents[0].Description, "devId3 - Låg Batterinivå")
	is.Equal(incRep.incidents[0].Category, 17)
}

func TestThatDeviceUpdatedSendsIncidentReportEvenOnUnknownState(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", "0")
	is.NoErr(err)

	err = app.DeviceStateUpdated("urn:ngsi-ld:Device:se:servanet:lora:msva:devId4", "3")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(1))
	is.Equal(incRep.incidents[0].Description, "devId4 - Okänt Fel: 3")
	is.Equal(incRep.incidents[0].Category, 17)
}

func testSetup(t *testing.T) (*is.I, *incidentReporter, IntegrationIncident) {
	is := is.New(t)
	incRep := newIncidentReporterThatReturns(nil)
	app := NewApplication(log.Logger, incRep.f, "", "")

	return is, incRep, app
}
