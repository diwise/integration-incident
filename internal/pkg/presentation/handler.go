package presentation

import (
	"compress/flate"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/diwise/integration-incident/internal/pkg/application"
	"github.com/diwise/integration-incident/internal/pkg/presentation/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func CreateRouterAndStartServing(log zerolog.Logger, app application.IntegrationIncident, servicePort string) error {
	r := chi.NewRouter()

	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	r.Use(compressor.Handler)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/api/notify", notificationHandler(app))

	p, err := cloudevents.NewHTTP()
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to create protocol: %s", err.Error())
	}

	h, err := cloudevents.NewHTTPReceiveHandler(context.Background(), p, receive(app))
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

func receive(app application.IntegrationIncident) func(context.Context, cloudevents.Event) {
	return func(ctx context.Context, event cloudevents.Event) {
		if strings.EqualFold(event.Type(), "diwise.statusmessage") {
			statusMessage := struct {
				DeviceID   string `json:"deviceID"`
				StatusCode int    `json:"statusCode"`
				Timestamp  string `json:"timestamp"`
			}{}

			json.Unmarshal(event.Data(), &statusMessage)

			if strings.Contains(statusMessage.DeviceID, "se:servanet:lora:msva:") {
				app.DeviceStateUpdated(statusMessage.DeviceID, strconv.Itoa(statusMessage.StatusCode))
			}
		}
	}
}

func notificationHandler(app application.IntegrationIncident) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		notif := api.Notification{}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Err(err).Msg("failed to read request body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(bodyBytes, &notif)
		if err != nil {
			log.Err(err).Msg("failed to unmarshal request body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if notif.SubscriptionId == "" {
			log.Err(err).Msg("request body is not a valid notification")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		for _, n := range notif.Data {
			switch n.Type {
			case "Device":
				if n.DeviceState != nil && strings.Contains(n.Id, "se:servanet:lora:msva:") {
					app.DeviceStateUpdated(n.Id, n.DeviceState.Value)
				}
			case "Lifebuoy":
				if n.Status != nil {
					app.LifebuoyValueUpdated(n.Id, n.Status.Value)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	})
}
