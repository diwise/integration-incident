package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/diwise/context-broker/pkg/ngsild/types/entities"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

type EntityLocator interface {
	Locate(ctx context.Context, entityType, entityID string) (latitude, longitude float64, err error)
}

func NewEntityLocator(host, tenant string) (EntityLocator, error) {
	return &locator{
		host:   host,
		tenant: tenant,
	}, nil
}

type locator struct {
	host   string
	tenant string
}

var tracer = otel.Tracer("integration-incident/svcs/locator")

const DefaultBrokerTenant string = "default"

var httpClient = http.Client{
	Transport: otelhttp.NewTransport(http.DefaultTransport),
	Timeout:   10 * time.Second,
}

func (l *locator) Locate(ctx context.Context, entityType, entityID string) (latitude, longitude float64, err error) {

	ctx, span := tracer.Start(ctx, "locate")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	log := logging.GetFromContext(ctx)
	_, ctx, log = o11y.AddTraceIDToLoggerAndStoreInContext(span, log, ctx)

	var req *http.Request

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, l.host+"/ngsi-ld/v1/entities/"+entityID+"?options=keyValues", nil)
	if err != nil {
		err = fmt.Errorf("failed to create request: %w", err)
		return
	}

	req.Header.Add("Accept", "application/ld+json")
	req.Header.Add("Link", entities.LinkHeader)

	if l.tenant != DefaultBrokerTenant {
		req.Header.Add("NGSILD-Tenant", l.tenant)
	}

	log.Info(fmt.Sprintf("requesting entity details for %s %s from %s", entityType, entityID, l.host))

	response, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("request failed: %w", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("request failed: %d != %d", response.StatusCode, http.StatusOK)
		return
	}

	entity := struct {
		Location struct {
			Type        string    `json:"type"`
			Coordinates []float64 `json:"coordinates"`
		} `json:"location"`
	}{}

	var b []byte
	b, _ = io.ReadAll(response.Body)
	err = json.Unmarshal(b, &entity)

	if err != nil {
		err = fmt.Errorf("failed to unmarshal entity: %w", err)
		return
	}

	if entity.Location.Type != "Point" {
		err = fmt.Errorf("entity location is missing or not a Point")
		return
	}

	latitude = entity.Location.Coordinates[1]
	longitude = entity.Location.Coordinates[0]

	return
}
