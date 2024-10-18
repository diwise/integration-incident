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

	log := logging.GetFromContext(ctx)

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodPost, gatewayUrl+"/token", body)
	if err != nil {
		err = fmt.Errorf("failed to create post request: %w", err)
		log.Error("request error", "err", err.Error())
		return nil, err
	}

	req.Header.Set("Authorization", authCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	resp, err = httpClient.Do(req)
	if err != nil {
		log.Error("request failed", "err", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid response %d from token endpoint", resp.StatusCode)
		log.Error("bad response", "err", err.Error())
		return nil, err
	}

	var bodyBytes []byte
	bodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response body (%w)", err)
		log.Error("i/o error", "err", err.Error())
		return nil, err
	}

	token := tokenResponse{}

	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		log.Error("failed to unmarshal access token json", "err", err.Error())
		return nil, err
	}

	log.Info("refreshed access token")

	return &token, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
