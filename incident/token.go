package incident

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/diwise/integration-incident/infrastructure/logging"
)

func getAccessToken(log logging.Logger, gatewayUrl, authCode string) (*tokenResponse, error) {
	params := url.Values{}
	params.Add("grant_type", `client_credentials`)
	body := strings.NewReader(params.Encode())

	req, err := http.NewRequest("POST", gatewayUrl+"/token", body)
	if err != nil {
		log.Errorf("failed to create post request: %s", err.Error())
		return nil, err
	}

	req.Header.Set("Authorization", authCode)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("failed to create get request: %s", err.Error())
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("invalid response: %s", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response body: %s", err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	token := tokenResponse{}

	err = json.Unmarshal(bodyBytes, &token)
	if err != nil {
		log.Errorf("failed to unmarshal access token json: %s", err.Error())
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
