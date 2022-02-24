package incident

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog"
)

func getAccessToken(log zerolog.Logger, gatewayUrl, authCode string) (*tokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", gatewayUrl+"/token", body)
	if err != nil {
		log.Err(err).Msg("failed to create post request")
		return nil, err
	}

	req.Header.Set("Authorization", authCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Err(err).Msg("failed to create get request")
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Err(nil).Msgf("invalid response: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("failed to read response body")
		return nil, err
	}

	defer resp.Body.Close()

	token := tokenResponse{}

	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal access token json")
		return nil, err
	}

	return &token, nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
