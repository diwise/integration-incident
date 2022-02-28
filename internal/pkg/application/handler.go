package application

import (
	"compress/flate"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/diwise/integration-incident/presentation/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func CreateRouterAndStartServing(log zerolog.Logger, servicePort string) error {
	r := chi.NewRouter()

	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		Debug:            false,
	}).Handler)

	compressor := middleware.NewCompressor(flate.DefaultCompression, "application/json", "application/ld+json")
	r.Use(compressor.Handler)
	r.Use(middleware.Logger)

	r.Post("/notification", notificationHandler())

	log.Info().Str("port", servicePort).Msg("starting to listen for connections")

	log.Log().Str("Starting integration on port:%s", servicePort)
	err := http.ListenAndServe(":"+servicePort, r)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen for connections")
	}

	return nil
}

func notificationHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		notif := api.Notification{}

		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Err(err).Msg("failed to read request body")
		}

		err = json.Unmarshal(bodyBytes, &notif)
		if err != nil {
			log.Err(err).Msg("failed to unmarshal request body")
			w.WriteHeader(http.StatusInternalServerError)
		}

		//check if we have already looked at device state, and if yes, what was previous state?

		if len(notif.Data) != 0 {
			for _, device := range notif.Data {
				if strings.Contains(device.ID, "se:servanet:lora:msva:") && device.DeviceState != nil {
					checkPreviousDeviceState(device.ID, device.DeviceState.Value)

				}
			}
		}
	})
}

func checkPreviousDeviceState(deviceId, state string) {
	_, exists := previousState[deviceId]

	if !exists {
		previousState[deviceId] = state
	}

	if previousState[deviceId] != state {
		//create incident
	}
}

var previousState map[string]string = make(map[string]string)
