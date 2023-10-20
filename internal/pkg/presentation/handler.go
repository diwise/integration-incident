package presentation

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/integration-incident/internal/pkg/presentation/api"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("integration-incident/handlers")

func CreateRouterAndStartServing(ctx context.Context, app application.IntegrationIncident, servicePort string) error {
	r := chi.NewRouter()

	log := logging.GetFromContext(ctx)

	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	r.Use(otelchi.Middleware("integration-incident", otelchi.WithChiRoutes(r)))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/api/notify", notificationHandler(log, app))

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(context.Background(), p, receive(log, app))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create handler: %s", err.Error())
	}

	r.Post("/api/cloudevents", cloudeventReceiveHandler(h))

	log.Info().Str("port", servicePort).Msg("starting to listen for connections")

	err = http.ListenAndServe(":"+servicePort, r)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen for connections")
	}

	return nil
}

func cloudeventReceiveHandler(h *client.EventReceiver) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}

func receive(logger zerolog.Logger, app application.IntegrationIncident) func(context.Context, cloudevents.Event) {
	return func(ctx context.Context, event cloudevents.Event) {
		var err error

		ctx, span := tracer.Start(ctx, "handle-cloudevent")
		defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

		_, ctx, log := o11y.AddTraceIDToLoggerAndStoreInContext(span, logger, ctx)

		log.Debug().Str("event_type", event.Type()).Msg("received cloud event")

		eventType := strings.ToLower(event.Type())

		switch eventType {
		case "diwise.statusmessage":
			statusMessage := models.StatusMessage{}

			err = json.Unmarshal(event.Data(), &statusMessage)
			if err != nil {			
				log.Err(err).Msg("failed to unmarshal event")
				return
			}

			if strings.Contains(statusMessage.DeviceID, "se:servanet:lora:msva:") {
				ctx = logging.NewContextWithLogger(ctx, log.With().Str("device_id", statusMessage.DeviceID).Logger())
				err = app.DeviceStateUpdated(ctx, statusMessage.DeviceID, statusMessage)
				if err != nil {			
					log.Err(err).Msg("device status updated failed")
					return
				}
			}
		case "function.updated":
			functionUpdated := models.FunctionUpdated{}

			err = json.Unmarshal(event.Data(), &functionUpdated)
			if err != nil {
				log.Err(err).Msg("failed to unmarshal event")
				return 
			}

			log.Debug().Msgf("function.updated, type")

			if functionUpdated.Type == "stopwatch" && functionUpdated.SubType == "overflow" {
				err = app.SewerOverflow(ctx, functionUpdated)
			}
			if err != nil {
				log.Err(err).Msg("sewer overflow failed")
				return 
			}
		default:
			log.Info().Str("event_type", event.Type()).Msg("ignoring unknown type")
		}
	}
}

func notificationHandler(logger zerolog.Logger, app application.IntegrationIncident) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		ctx, span := tracer.Start(r.Context(), "handle-notification")
		defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

		_, ctx, log := o11y.AddTraceIDToLoggerAndStoreInContext(span, logger, ctx)

		notif := api.Notification{}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			err = fmt.Errorf("failed to read request body (%w)", err)
			log.Err(err).Msg("i/o error")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(bodyBytes, &notif)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal request body (%w)", err)
			log.Err(err).Msg("bad request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if notif.SubscriptionId == "" {
			err = fmt.Errorf("request body is not a valid notification")
			log.Err(err).Msg("bad request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, n := range notif.Data {
			switch n.Type {
			case "Device":
				if n.DeviceState != nil && strings.Contains(n.Id, "se:servanet:lora:msva:") {
					code, _ := strconv.Atoi(n.DeviceState.Value)
					s := models.NewStatusMessage(n.Id, code)
					// TODO: remove code block?
					err = app.DeviceStateUpdated(ctx, n.Id, s)
				}
			case "Lifebuoy":
				if n.Status != nil {
					err = app.LifebuoyValueUpdated(ctx, n.Id, n.Status.Value)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	})
}
