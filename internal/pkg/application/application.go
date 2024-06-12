package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diwise/integration-incident/internal/pkg/application/services"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/integration-incident/pkg/incident"
	"github.com/diwise/ngsi-ld-golang/pkg/datamodels/diwise"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"github.com/rs/zerolog/log"
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

	deviceState := strconv.Itoa(sm.Code)
	if deviceState == PayloadError {
		log.Warn().Msg("ignoring payload error")
		return nil
	}

	key := fmt.Sprintf("%s:%s", shortId, "state")
	exists, changed := a.cache.ExistsAndIsChanged(key, deviceState)

	if exists && !changed {
		return nil
	}

	log.Info().Msgf("device state changed to %s", deviceState)

	if deviceState != StateNoError {
		const watermeterCategory int = 17
		errorType := Join(sm.Messages, " ", translate)

		if !strings.Contains(errorType, "Spricka") && !strings.Contains(errorType, "Läckage") && !strings.Contains(errorType, "Is") {
			log.Info().Msgf("device state contains error of type '%s', which is not prioritised", errorType)
			return nil
		}

		if errorType == "Is eller Frys Varning" && !withinBounds(sm.Timestamp) {
			log.Info().Msg("a freeze warning was received, but timestamp is out of bounds")
			return nil
		}

		incident := models.NewIncident(watermeterCategory, translateJoin(deviceId, sm)).AtLocation(62.388178, 17.315090)

		err := a.incidentReporter(ctx, *incident)
		if err != nil {
			log.Error().Err(err).Msg("could not post incident")
			return err
		}
	}

	a.cache.Add(key, deviceState)
	return nil
}

func withinBounds(timestamp string) bool {
	parsed, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.Error().Err(err).Msg("could not parse timestamp")
		return false
	}

	if parsed.Month() > 4 && parsed.Month() < 10 {
		return false
	}

	return true
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

	key := fmt.Sprintf("%s:%s", shortId, "value")
	exists, changed := a.cache.ExistsAndIsChanged(key, deviceValue)

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

		log.Info().Msgf("SewageOverflowObserved, id: %s, name: %s", functionUpdated.Id, functionUpdated.Name)

		incident := models.NewIncident(SewageOverflowObservedCategory, fmt.Sprintf("Bräddning upptäckt vid %s", functionUpdated.Name))

		if functionUpdated.Location != nil {
			incident = incident.AtLocation(functionUpdated.Location.Latitude, functionUpdated.Location.Longitude)
		}

		err = a.incidentReporter(ctx, *incident)
		if err != nil {
			log.Err(err).Msg("could not post incident")
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
