// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package application

import (
	"context"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
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
// 			DeviceStateUpdatedFunc: func(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error {
// 				panic("mock out the DeviceStateUpdated method")
// 			},
// 			LifebuoyValueUpdatedFunc: func(ctx context.Context, deviceId string, deviceValue string) error {
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
	DeviceStateUpdatedFunc func(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error

	// LifebuoyValueUpdatedFunc mocks the LifebuoyValueUpdated method.
	LifebuoyValueUpdatedFunc func(ctx context.Context, deviceId string, deviceValue string) error

	// calls tracks calls to the methods.
	calls struct {
		// DeviceStateUpdated holds details about calls to the DeviceStateUpdated method.
		DeviceStateUpdated []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// DeviceId is the deviceId argument value.
			DeviceId string
			// StatusMessage is the statusMessage argument value.
			StatusMessage models.StatusMessage
		}
		// LifebuoyValueUpdated holds details about calls to the LifebuoyValueUpdated method.
		LifebuoyValueUpdated []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
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
func (mock *IntegrationIncidentMock) DeviceStateUpdated(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error {
	if mock.DeviceStateUpdatedFunc == nil {
		panic("IntegrationIncidentMock.DeviceStateUpdatedFunc: method is nil but IntegrationIncident.DeviceStateUpdated was just called")
	}
	callInfo := struct {
		Ctx           context.Context
		DeviceId      string
		StatusMessage models.StatusMessage
	}{
		Ctx:           ctx,
		DeviceId:      deviceId,
		StatusMessage: statusMessage,
	}
	mock.lockDeviceStateUpdated.Lock()
	mock.calls.DeviceStateUpdated = append(mock.calls.DeviceStateUpdated, callInfo)
	mock.lockDeviceStateUpdated.Unlock()
	return mock.DeviceStateUpdatedFunc(ctx, deviceId, statusMessage)
}

// DeviceStateUpdatedCalls gets all the calls that were made to DeviceStateUpdated.
// Check the length with:
//     len(mockedIntegrationIncident.DeviceStateUpdatedCalls())
func (mock *IntegrationIncidentMock) DeviceStateUpdatedCalls() []struct {
	Ctx           context.Context
	DeviceId      string
	StatusMessage models.StatusMessage
} {
	var calls []struct {
		Ctx           context.Context
		DeviceId      string
		StatusMessage models.StatusMessage
	}
	mock.lockDeviceStateUpdated.RLock()
	calls = mock.calls.DeviceStateUpdated
	mock.lockDeviceStateUpdated.RUnlock()
	return calls
}

// LifebuoyValueUpdated calls LifebuoyValueUpdatedFunc.
func (mock *IntegrationIncidentMock) LifebuoyValueUpdated(ctx context.Context, deviceId string, deviceValue string) error {
	if mock.LifebuoyValueUpdatedFunc == nil {
		panic("IntegrationIncidentMock.LifebuoyValueUpdatedFunc: method is nil but IntegrationIncident.LifebuoyValueUpdated was just called")
	}
	callInfo := struct {
		Ctx         context.Context
		DeviceId    string
		DeviceValue string
	}{
		Ctx:         ctx,
		DeviceId:    deviceId,
		DeviceValue: deviceValue,
	}
	mock.lockLifebuoyValueUpdated.Lock()
	mock.calls.LifebuoyValueUpdated = append(mock.calls.LifebuoyValueUpdated, callInfo)
	mock.lockLifebuoyValueUpdated.Unlock()
	return mock.LifebuoyValueUpdatedFunc(ctx, deviceId, deviceValue)
}

// LifebuoyValueUpdatedCalls gets all the calls that were made to LifebuoyValueUpdated.
// Check the length with:
//     len(mockedIntegrationIncident.LifebuoyValueUpdatedCalls())
func (mock *IntegrationIncidentMock) LifebuoyValueUpdatedCalls() []struct {
	Ctx         context.Context
	DeviceId    string
	DeviceValue string
} {
	var calls []struct {
		Ctx         context.Context
		DeviceId    string
		DeviceValue string
	}
	mock.lockLifebuoyValueUpdated.RLock()
	calls = mock.calls.LifebuoyValueUpdated
	mock.lockLifebuoyValueUpdated.RUnlock()
	return calls
}
