package incident

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
	"github.com/rs/zerolog"
)

var errNotAuthorized = errors.New("invalid auth code or token refresh required")

func NewIncidentReporter(log zerolog.Logger, gatewayUrl, authCode string) (func(models.Incident) error, error) {
	token, err := getAccessToken(log, gatewayUrl, authCode)
	if err != nil {
		return nil, err
	}
	return func(incident models.Incident) error {
		err := postIncident(log, incident, gatewayUrl, token.AccessToken)
		if err == errNotAuthorized {
			log.Err(err).Msg("post incident failed, retrying with refreshed access token")
			token, _ = getAccessToken(log, gatewayUrl, authCode)
			return postIncident(log, incident, gatewayUrl, token.AccessToken)
		}
		return err
	}, nil
}

func postIncident(log zerolog.Logger, incident models.Incident, gatewayUrl, token string) error {

	incidentBytes, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("could not marshal incident message into json: %s", err.Error())
	}

	gatewayUrl = gatewayUrl + "/incident/1.0/api/sendincident"

	log.Info().Msgf("posting incident \"%s\" (cat: %d) to: %s", incident.Description, incident.Category, gatewayUrl)

	client := http.Client{}

	req, _ := http.NewRequest("POST", gatewayUrl, bytes.NewBuffer(incidentBytes))
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to post incident message: %s", err.Error())
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errNotAuthorized
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response code from backend: %d", resp.StatusCode)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err.Error())
	}

	response := incidentResponse{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal incident response: %s", err.Error())
	}

	if response.Status != "OK" {
		return fmt.Errorf("incident backend returned status \"%s\" with message \"%s\"", response.Status, response.Message)
	}

	log.Info().Msgf("incident created with ID: %s", response.IncidentID)

	return nil
}

type incidentResponse struct {
	Status     string `json:"status"`
	IncidentID string `json:"incidentId"`
	Message    string `json:"message"`
}
