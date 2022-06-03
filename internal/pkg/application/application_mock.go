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
// 			LifebuoyValueUpdatedFunc: func(deviceId string, deviceValue string) error {
// 				panic("mock out the LifebuoyValueUpdated method")
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

	// LifebuoyValueUpdatedFunc mocks the LifebuoyValueUpdated method.
	LifebuoyValueUpdatedFunc func(deviceId string, deviceValue string) error

	// calls tracks calls to the methods.
	calls struct {
		// DeviceStateUpdated holds details about calls to the DeviceStateUpdated method.
		DeviceStateUpdated []struct {
			// DeviceId is the deviceId argument value.
			DeviceId string
			// DeviceState is the deviceState argument value.
			DeviceState string
		}
		// LifebuoyValueUpdated holds details about calls to the LifebuoyValueUpdated method.
		LifebuoyValueUpdated []struct {
			// DeviceId is the deviceId argument value.
			DeviceId string
			// DeviceValue is the deviceValue argument value.
			DeviceValue string
		}
	}
	lockDeviceStateUpdated   sync.RWMutex
	lockLifebuoyValueUpdated sync.RWMutex
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

// LifebuoyValueUpdated calls LifebuoyValueUpdatedFunc.
func (mock *IntegrationIncidentMock) LifebuoyValueUpdated(deviceId string, deviceValue string) error {
	if mock.LifebuoyValueUpdatedFunc == nil {
		panic("IntegrationIncidentMock.LifebuoyValueUpdatedFunc: method is nil but IntegrationIncident.LifebuoyValueUpdated was just called")
	}
	callInfo := struct {
		DeviceId    string
		DeviceValue string
	}{
		DeviceId:    deviceId,
		DeviceValue: deviceValue,
	}
	mock.lockLifebuoyValueUpdated.Lock()
	mock.calls.LifebuoyValueUpdated = append(mock.calls.LifebuoyValueUpdated, callInfo)
	mock.lockLifebuoyValueUpdated.Unlock()
	return mock.LifebuoyValueUpdatedFunc(deviceId, deviceValue)
}

// LifebuoyValueUpdatedCalls gets all the calls that were made to LifebuoyValueUpdated.
// Check the length with:
//     len(mockedIntegrationIncident.LifebuoyValueUpdatedCalls())
func (mock *IntegrationIncidentMock) LifebuoyValueUpdatedCalls() []struct {
	DeviceId    string
	DeviceValue string
} {
	var calls []struct {
		DeviceId    string
		DeviceValue string
	}
	mock.lockLifebuoyValueUpdated.RLock()
	calls = mock.calls.LifebuoyValueUpdated
	mock.lockLifebuoyValueUpdated.RUnlock()
	return calls
}
