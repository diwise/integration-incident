package incident

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"slices"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var errNotAuthorized = errors.New("invalid auth code or token refresh required")

type ReporterFunc func(context.Context, models.Incident) error

func NewIncidentReporter(ctx context.Context, gatewayUrl, authCode string) (ReporterFunc, error) {
	token, err := getAccessToken(ctx, gatewayUrl, authCode)
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, incident models.Incident) error {
		err := postIncident(ctx, incident, gatewayUrl, token.AccessToken)
		if err == errNotAuthorized {
			log := logging.GetFromContext(ctx)
			log.Error().Err(err).Msg("post incident failed, retrying after access token refresh")

			newToken, err := getAccessToken(ctx, gatewayUrl, authCode)
			if err != nil {
				err = fmt.Errorf("failed to refresh access token: %w", err)
				return err
			}

			token = newToken
			return postIncident(ctx, incident, gatewayUrl, token.AccessToken)
		}
		return err
	}, nil
}

func postIncident(ctx context.Context, incident models.Incident, gatewayUrl, token string) error {
	var err error
	ctx, span := tracer.Start(ctx, "post-incident")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	var incidentBytes []byte
	incidentBytes, err = json.Marshal(incident)
	if err != nil {
		err = fmt.Errorf("could not marshal incident message into json: %w", err)
		return err
	}

	gatewayUrl = gatewayUrl + "/incident/2.0/incident"

	log := logging.GetFromContext(ctx)
	log.Info().Msgf("posting incident \"%s\" (cat: %d) to: %s", incident.Description, incident.Category, gatewayUrl)

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, gatewayUrl, bytes.NewBuffer(incidentBytes))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Error().Err(err).Msg("Could not dump the request")
	} else {
		log.Debug().Msgf("HTTP request: %s", dump)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to post incident message: %w", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		err = errNotAuthorized
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad response code from backend: %d", resp.StatusCode)
		return err
	}

	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body: %w", err)
		return err
	}

	response := incidentResponse{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal incident response: %w", err)
		return err
	}

	if !slices.Contains([]string{"SPARAT", "INSKICKAT", "KLART"}, response.Status) {
		err = fmt.Errorf("incident backend returned status \"%s\" with message \"%s\"", response.Status, response.Message)
		return err
	}

	log.Info().Msgf("incident created with ID: %s", response.IncidentID)

	return nil
}

type incidentResponse struct {
	Status     string `json:"status"`
	IncidentID string `json:"incidentId"`
	Message    string `json:"message"`
}
