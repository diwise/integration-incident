package application

import (
	"testing"

	"github.com/matryer/is"
	"github.com/rs/zerolog"
)

func TestThatDeviceStateUpdatedDoesSomething(t *testing.T) {
	is, _, app := testSetup(t)

	err := app.DeviceStateUpdated("devId", "devState")
	is.NoErr(err)
}

func TestThatDeviceStateUpdatedSendsIncidentReportOnStateChanged(t *testing.T) {
	is, incRep, app := testSetup(t)

	err := app.DeviceStateUpdated("devId", "devState")
	is.NoErr(err)

	err = app.DeviceStateUpdated("devId", "notthesame")
	is.NoErr(err)
	is.Equal(incRep.callCount, int32(1))
	is.Equal(incRep.incidents[0].Description, "devId - notthesame")
	is.Equal(incRep.incidents[0].Category, 16)
}

func testSetup(t *testing.T) (*is.I, *incidentReporter, IntegrationIncident) {
	is := is.New(t)
	log := zerolog.Logger{}
	incRep := newIncidentReporterThatReturns(nil)
	app := NewApplication(log, incRep.f, "", "")

	return is, incRep, app
}
