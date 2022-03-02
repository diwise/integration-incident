package application

import (
	"testing"

	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceDoesNotExist(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("devId1", "devState")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))
}

func TestThatDeviceStateUpdatedDoesNotSendIncidentIfDeviceStateIsTheSame(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("devId2", "moto")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))

	err = app.DeviceStateUpdated("devId2", "moto")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(0))
}

func TestThatDeviceStateUpdatedSendsIncidentReportOnStateChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("devId3", "devState")
	is.NoErr(err)

	err = app.DeviceStateUpdated("devId3", "notthesame")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(1))
	is.Equal(incRep.incidents[0].Description, "devId3 - notthesame")
	is.Equal(incRep.incidents[0].Category, 16)
}

func testSetup(t *testing.T) (*is.I, *incidentReporter, IntegrationIncident) {
	is := is.New(t)
	log := zerolog.Logger{}
	incRep := newIncidentReporterThatReturns(nil)
	app := NewApplication(log, incRep.f, "", "")

	return is, incRep, app
}
