package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/otel"
)

//go:generate moq -rm -out application_mock.go . IntegrationIncident

type IntegrationIncident interface {
	DeviceStateUpdated(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error
	LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error
	SewerOverflow(ctx context.Context, functionUpdated models.FunctionUpdated) error
}

var tracer = otel.Tracer("integration-incident/app")

type app struct {
	incidentReporter incident.ReporterFunc
	entityLocator    services.EntityLocator
	stateMutex       sync.Mutex
	previousStates   map[string]string
	previousValues   map[string]string
}

func NewApplication(_ context.Context, incidentReporter incident.ReporterFunc, entityLocator services.EntityLocator) IntegrationIncident {

	newApp := &app{
		incidentReporter: incidentReporter,
		entityLocator:    entityLocator,
		previousStates:   make(map[string]string),
		previousValues:   make(map[string]string),
	}

	return newApp
}

func (a *app) DeviceStateUpdated(ctx context.Context, deviceId string, sm models.StatusMessage) error {
	var err error

	if !strings.Contains(deviceId, "se:servanet:lora:msva:") {
		return fmt.Errorf("device with id %s is not supported", deviceId)
	}

	log := logging.GetFromContext(ctx)

	ctx, span := tracer.Start(ctx, "device-state-updated")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, log, ctx)

	shortId := deviceId[strings.LastIndex(deviceId, ":")+1:]

	const (
		StateNoError string = "0"
		PayloadError string = "100"
	)

	deviceState := strconv.Itoa(sm.Code)
	if deviceState == PayloadError {
		log.Warn().Msg("ignoring payload error")
		return nil
	}

	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	exists, changed := a.checkIfDeviceStateExistsAndHasChanged(shortId, deviceState)
	if exists && !changed {
		return nil
	}

	log.Info().Msgf("device state changed to %s", deviceState)

	if deviceState != StateNoError {
		const watermeterCategory int = 17
		incident := models.NewIncident(watermeterCategory, translateJoin(shortId, sm)).AtLocation(62.388178, 17.315090)

		err := a.incidentReporter(ctx, *incident)
		if err != nil {
			log.Error().Err(err).Msg("could not post incident")
			return err
		}
	}

	a.updateDeviceState(shortId, deviceState)

	return nil
}

func (a *app) LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error {
	var err error

	ctx, span := tracer.Start(ctx, "lifebuoy-updated")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	log := logging.GetFromContext(ctx)
	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, log, ctx)

	if !strings.HasPrefix(deviceId, diwise.LifebuoyIDPrefix) {
		err = fmt.Errorf("device with id %s is not supported", deviceId)
		return err
	}

	shortId := strings.TrimPrefix(deviceId, diwise.LifebuoyIDPrefix)

	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()

	exists, changed := a.checkIfDeviceValueExistsAndHasChanged(shortId, deviceValue)

	if exists && !changed {
		return nil
	}

	if deviceValue == "off" {
		log.Info().Msgf("state changed to \"off\" on device: %s", shortId)

		const lifebuoyCategory int = 15
		incident := models.NewIncident(lifebuoyCategory, "Livboj kan ha flyttats eller utsatts för åverkan.")

		latitude, longitude, err := a.entityLocator.Locate(ctx, diwise.LifebuoyTypeName, deviceId)
		if err == nil {
			incident = incident.AtLocation(latitude, longitude)
		}

		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			log.Err(err).Msg("could not post incident")
			return err
		}
	}

	a.updateDeviceValue(shortId, deviceValue)

	return nil
}

func (a *app) SewerOverflow(ctx context.Context, functionUpdated models.FunctionUpdated) error {
	var err error

	ctx, span := tracer.Start(ctx, "sewer-overflow-detected")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	log := logging.GetFromContext(ctx)
	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, log, ctx)

	if functionUpdated.Stopwatch.State {
		const SewerOverflowCategory int = 18

		log.Info().Msgf("Sewer overflow detected, id: %s, name: %s", functionUpdated.Id, functionUpdated.Name)

		incident := models.NewIncident(SewerOverflowCategory, fmt.Sprintf("Bräddning upptäckt vid %s", functionUpdated.Name))
		
		if functionUpdated.Location != nil {
			incident = incident.AtLocation(functionUpdated.Location.Latitude, functionUpdated.Location.Longitude)
		}
		
		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			log.Err(err).Msg("could not post incident")
			return err
		}
	}

	return nil
}

func (a *app) updateDeviceState(deviceId, deviceState string) {
	a.previousStates[deviceId] = deviceState
}

func (a *app) updateDeviceValue(deviceId, deviceValue string) {
	a.previousValues[deviceId] = deviceValue
}

func (a *app) checkIfDeviceStateExistsAndHasChanged(deviceId, state string) (exists, changed bool) {
	var storedState string

	storedState, exists = a.previousStates[deviceId]

	if !exists {
		a.previousStates[deviceId] = state
	} else if storedState != state {
		changed = true
	}

	return
}

func (a *app) checkIfDeviceValueExistsAndHasChanged(deviceId, value string) (exists, changed bool) {
	var storedValue string

	storedValue, exists = a.previousValues[deviceId]

	if !exists {
		a.previousValues[deviceId] = value
	} else if storedValue != value {
		changed = true
	}

	return
}

func translateJoin(deviceID string, sm models.StatusMessage) string {
	return fmt.Sprintf("%s - %s", deviceID, Join(sm.Messages, " ", translate))
}

func Join(elems []string, sep string, mod func(string) string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return mod(elems[0])
	}
	n := len(sep) * (len(elems) - 1)
	for i := 0; i < len(elems); i++ {
		n += len(elems[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(mod(elems[0]))
	for _, s := range elems[1:] {
		b.WriteString(sep)
		b.WriteString(mod(s))
	}
	return b.String()
}

func translate(s string) string {
	switch s {
	case "No error":
		return "Inga fel"
	case "Power low":
		return "Låg batterinivå"
	case "Permanent error":
		return "Permanent fel"
	case "Temporary error":
		return "Temporärt fel"
	case "Empty spool":
		return "Tomt rör"
	case "Leak":
		return "Läckage"
	case "Burst":
		return "Spricka"
	case "Backflow":
		return "Backflöde"
	case "Freeze":
		return "Is eller Frys Varning"
	case "Unknown":
		return "Okänt fel"
	}
	return s
}
