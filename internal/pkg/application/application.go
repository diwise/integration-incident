package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/otel"
)

//go:generate moq -rm -out application_mock.go . IntegrationIncident

type IntegrationIncident interface {
	DeviceStateUpdated(ctx context.Context, deviceId string, statusMessage models.StatusMessage) error
	LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error
	SewageOverflowObserved(ctx context.Context, functionUpdated models.FunctionUpdated) error
}

var tracer = otel.Tracer("integration-incident/app")

type cache struct {
	mx    sync.Mutex
	items map[string]string
}

func (c *cache) Add(key, value string) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items[key] = value
}
func (c *cache) Equals(key, value string) bool {
	c.mx.Lock()
	defer c.mx.Unlock()
	storedValue, ok := c.items[key]
	if !ok {
		return false
	}
	return storedValue == value
}
func (c *cache) ExistsAndIsChanged(key, value string) (bool, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()
	storedValue, ok := c.items[key]
	if !ok {
		return false, false
	}
	return true, storedValue != value
}

type app struct {
	incidentReporter incident.ReporterFunc
	entityLocator    services.EntityLocator
	cache            cache
}

func NewApplication(_ context.Context, incidentReporter incident.ReporterFunc, entityLocator services.EntityLocator) IntegrationIncident {

	newApp := &app{
		incidentReporter: incidentReporter,
		entityLocator:    entityLocator,
		cache:            cache{items: make(map[string]string)},
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

	if sm.Code == nil {
		b, _ := json.Marshal(sm)
		log.Debug("statusCode did not contain any information", "device_id", deviceId, "body", string(b))
		return nil
	}

	deviceState := *sm.Code

	if deviceState == PayloadError {
		log.Warn("ignoring payload error")
		return nil
	}

	if deviceState == StateNoError {
		return nil
	}

	key := fmt.Sprintf("%s:%s", shortId, "state")
	exists, changed := a.cache.ExistsAndIsChanged(key, deviceState)

	if exists && !changed {
		log.Debug("device state has not changed", "device_id", deviceId, "state", deviceState)
		return nil
	}

	log.Info("device state changed", "state", deviceState)

	const watermeterCategory int = 17
	errorType := Join(sm.Messages, " ", translate)

	if !strings.Contains(errorType, "Spricka") && !strings.Contains(errorType, "Läckage") && !strings.Contains(errorType, "Is") {
		log.Info(fmt.Sprintf("device state contains error of type '%s', which is not prioritised", errorType))
		return nil
	}

	if errorType == "Is eller Frys Varning" && !withinBounds(ctx, sm.Timestamp) {
		log.Info("a freeze warning was received, but timestamp is out of bounds")
		return nil
	}

	incident := models.NewIncident(watermeterCategory, translateJoin(deviceId, sm)).AtLocation(62.388178, 17.315090)

	err = a.incidentReporter(ctx, *incident)
	if err != nil {
		err = fmt.Errorf("could not post incident: %s", err.Error())
		return err
	}

	a.cache.Add(key, deviceState)
	return nil
}

func withinBounds(_ context.Context, timestamp time.Time) bool {
	if timestamp.Month() > 4 && timestamp.Month() < 10 {
		return false
	}

	return true
}

func (a *app) LifebuoyValueUpdated(ctx context.Context, deviceId, deviceValue string) error {
	var err error
	var log *slog.Logger

	ctx, span := tracer.Start(ctx, "lifebuoy-updated")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, logging.GetFromContext(ctx), ctx)

	const LifebuoyTypeName string = "Lifebuoy"
	const LifebuoyIDPrefix string = "urn:ngsi-ld:" + LifebuoyTypeName + ":"

	if !strings.HasPrefix(deviceId, LifebuoyIDPrefix) {
		err = fmt.Errorf("device with id %s is not supported", deviceId)
		return err
	}

	shortId := strings.TrimPrefix(deviceId, LifebuoyIDPrefix)

	key := fmt.Sprintf("%s:%s", shortId, "value")
	exists, changed := a.cache.ExistsAndIsChanged(key, deviceValue)

	if exists && !changed {
		return nil
	}

	if deviceValue == "off" {
		log.Info("state changed to \"off\" on device", "device_id", shortId)

		const lifebuoyCategory int = 15
		incident := models.NewIncident(lifebuoyCategory, "Livboj kan ha flyttats eller utsatts för åverkan.")

		latitude, longitude, err := a.entityLocator.Locate(ctx, LifebuoyTypeName, deviceId)
		if err == nil {
			incident = incident.AtLocation(latitude, longitude)
		}

		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			err = fmt.Errorf("could not post incident: %s", err.Error())
			return err
		}
	}

	a.cache.Add(key, deviceValue)

	return nil
}

func (a *app) SewageOverflowObserved(ctx context.Context, functionUpdated models.FunctionUpdated) error {
	var err error

	ctx, span := tracer.Start(ctx, "sewage-overflow-observed")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	log := logging.GetFromContext(ctx)
	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, log, ctx)

	key := fmt.Sprintf("%s:%s:%s", functionUpdated.Id, functionUpdated.Type, functionUpdated.SubType)

	if a.cache.Equals(key, strconv.FormatBool(functionUpdated.Stopwatch.State)) {
		return nil
	}

	if functionUpdated.Stopwatch.State {
		const SewageOverflowObservedCategory int = 18

		log.Info(fmt.Sprintf("SewageOverflowObserved, id: %s, name: %s", functionUpdated.Id, functionUpdated.Name))

		incident := models.NewIncident(SewageOverflowObservedCategory, fmt.Sprintf("Bräddning upptäckt vid %s", functionUpdated.Name))

		if functionUpdated.Location != nil {
			incident = incident.AtLocation(functionUpdated.Location.Latitude, functionUpdated.Location.Longitude)
		}

		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			err = fmt.Errorf("could not post incident: %s", err.Error())
			return err
		}
	}

	a.cache.Add(key, strconv.FormatBool(functionUpdated.Stopwatch.State))

	return nil
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
	for i := range elems {
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
