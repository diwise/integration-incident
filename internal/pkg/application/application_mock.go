// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package application

import (
	"sync"
)

// Ensure, that IntegrationIncidentMock does implement IntegrationIncident.
// If this is not the case, regenerate this file with moq.
var _ IntegrationIncident = &IntegrationIncidentMock{}

// IntegrationIncidentMock is a mock implementation of IntegrationIncident.
//
// 	func TestSomethingThatUsesIntegrationIncident(t *testing.T) {
//
// 		// make and configure a mocked IntegrationIncident
// 		mockedIntegrationIncident := &IntegrationIncidentMock{
// 			DeviceStateUpdatedFunc: func(deviceId string, deviceState string) error {
// 				panic("mock out the DeviceStateUpdated method")
// 			},
// 			StartFunc: func() error {
// 				panic("mock out the Start method")
// 			},
// 		}
//
// 		// use mockedIntegrationIncident in code that requires IntegrationIncident
// 		// and then make assertions.
//
// 	}
type IntegrationIncidentMock struct {
	// DeviceStateUpdatedFunc mocks the DeviceStateUpdated method.
	DeviceStateUpdatedFunc func(deviceId string, deviceState string) error

	// StartFunc mocks the Start method.
	StartFunc func() error

	// calls tracks calls to the methods.
	calls struct {
		// DeviceStateUpdated holds details about calls to the DeviceStateUpdated method.
		DeviceStateUpdated []struct {
			// DeviceId is the deviceId argument value.
			DeviceId string
			// DeviceState is the deviceState argument value.
			DeviceState string
		}
		// Start holds details about calls to the Start method.
		Start []struct {
		}
	}
	lockDeviceStateUpdated sync.RWMutex
	lockStart              sync.RWMutex
}

// DeviceStateUpdated calls DeviceStateUpdatedFunc.
func (mock *IntegrationIncidentMock) DeviceStateUpdated(deviceId string, deviceState string) error {
	if mock.DeviceStateUpdatedFunc == nil {
		panic("IntegrationIncidentMock.DeviceStateUpdatedFunc: method is nil but IntegrationIncident.DeviceStateUpdated was just called")
	}
	callInfo := struct {
		DeviceId    string
		DeviceState string
	}{
		DeviceId:    deviceId,
		DeviceState: deviceState,
	}
	mock.lockDeviceStateUpdated.Lock()
	mock.calls.DeviceStateUpdated = append(mock.calls.DeviceStateUpdated, callInfo)
	mock.lockDeviceStateUpdated.Unlock()
	return mock.DeviceStateUpdatedFunc(deviceId, deviceState)
}

// DeviceStateUpdatedCalls gets all the calls that were made to DeviceStateUpdated.
// Check the length with:
//     len(mockedIntegrationIncident.DeviceStateUpdatedCalls())
func (mock *IntegrationIncidentMock) DeviceStateUpdatedCalls() []struct {
	DeviceId    string
	DeviceState string
} {
	var calls []struct {
		DeviceId    string
		DeviceState string
	}
	mock.lockDeviceStateUpdated.RLock()
	calls = mock.calls.DeviceStateUpdated
	mock.lockDeviceStateUpdated.RUnlock()
	return calls
}

// Start calls StartFunc.
func (mock *IntegrationIncidentMock) Start() error {
	if mock.StartFunc == nil {
		panic("IntegrationIncidentMock.StartFunc: method is nil but IntegrationIncident.Start was just called")
	}
	callInfo := struct {
	}{}
	mock.lockStart.Lock()
	mock.calls.Start = append(mock.calls.Start, callInfo)
	mock.lockStart.Unlock()
	return mock.StartFunc()
}

// StartCalls gets all the calls that were made to Start.
// Check the length with:
//     len(mockedIntegrationIncident.StartCalls())
func (mock *IntegrationIncidentMock) StartCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockStart.RLock()
	calls = mock.calls.Start
	mock.lockStart.RUnlock()
	return calls
}
