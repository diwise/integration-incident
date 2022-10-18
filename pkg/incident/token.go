package incident

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/logging"
	"github.com/diwise/service-chassis/pkg/infrastructure/o11y/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("integration-incident/token")

func getAccessToken(ctx context.Context, gatewayUrl, authCode string) (*tokenResponse, error) {
	var err error
	ctx, span := tracer.Start(ctx, "token-refresh")
	defer func() { tracing.RecordAnyErrorAndEndSpan(err, span) }()

	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	body := strings.NewReader(params.Encode())

	httpClient := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	log := logging.GetFromContext(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, gatewayUrl+"/token", body)
	if err != nil {
		err = fmt.Errorf("failed to create post request: %w", err)
		log.Err(err).Msg("request error")
		return nil, err
	}

	req.Header.Set("Authorization", authCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Err(err).Msg("request failed")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid response %d from token endpoint", resp.StatusCode)
		log.Error().Err(err).Msgf("bad response")
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body (%w)", err)
		log.Err(err).Msg("i/o error")
		return nil, err
	}

	token := tokenResponse{}

	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal access token json")
		return nil, err
	}

	log.Info().Msg("refreshed access token")

	return &token, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
