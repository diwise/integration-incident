package presentation

import (
	"compress/flate"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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

	log.Info().Str("port", servicePort).Msg("starting to listen for connections")

	err := http.ListenAndServe(":"+servicePort, r)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen for connections")
	}

	return nil
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

		if len(notif.Data) != 0 {
			for _, r := range notif.Data {
				t := r["type"].(string)
				switch t {
				case "Device":
					err = handleDevice(app, r)
				case "Lifebuoy":
					err = handleLifebuoy(app, r)
				default:
					log.Warn().Msgf("could not handle type %s", t)
				}

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	})
}

func handleDevice(app application.IntegrationIncident, r map[string]interface{}) error {
	id, ok := getStringFromMap("id", r)
	if !ok {
		return fmt.Errorf("could not find id")
	}

	state, ok := getStringFromMap("deviceState", r)
	if !ok {
		return nil
	}

	if strings.Contains(id, "se:servanet:lora:msva:") && state != "" {
		err := app.DeviceStateUpdated(id, state)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleLifebuoy(app application.IntegrationIncident, r map[string]interface{}) error {
	id, ok := getStringFromMap("id", r)
	if !ok {
		return fmt.Errorf("could not find id")
	}

	status, ok := getStringFromMap("status", r)
	if !ok {
		return nil
	}

	if status != "" {
		err := app.LifebuoyValueUpdated(id, status)
		if err != nil {
			return err
		}
	}
	return nil
}

func getStringFromMap(key string, m map[string]interface{}) (string, bool) {
	if val, ok := m[key]; ok {
		if s, ok := val.(string); ok {
			return s, true
		}
	}
	return "", false
}
