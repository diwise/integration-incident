package incident

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/logging"
	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
)

func NewIncidentReporter(log logging.Logger, gatewayUrl, authCode string) (func(models.Incident) error, error) {
	token, err := getAccessToken(log, gatewayUrl, authCode)
	if err != nil {
		return nil, err
	}
	return func(incident models.Incident) error {
		err := postIncident(log, incident, gatewayUrl, token.AccessToken)
		if err != nil {
			log.Infof("post incident failed, retrying with refreshed access token: %s", err.Error())
			token, _ = getAccessToken(log, gatewayUrl, authCode)
			return postIncident(log, incident, gatewayUrl, token.AccessToken)
		}
		return err
	}, nil
}

func postIncident(log logging.Logger, incident models.Incident, gatewayUrl, token string) error {

	incidentBytes, err := json.Marshal(incident)
	if err != nil {
		log.Errorf("could not marshal incident message into json: %s", err.Error())
	}

	gatewayUrl = gatewayUrl + "/incident/v01/api/sendincident"

	log.Infof("posting incident to: %s", gatewayUrl)

	client := http.Client{}

	req, _ := http.NewRequest("POST", gatewayUrl, bytes.NewBuffer(incidentBytes))
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to post incident message: %s", err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("failed to create incident: %s", resp.StatusCode)
		return err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("failed to read response body: %s", err.Error())
	}

	response := incidentResponse{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		log.Error("failed to unmarshal incident response: %s", err.Error())
	}

	log.Infof("status ok, incident created with ID: %s", response.IncidentID)

	return nil
}

type incidentResponse struct {
	IncidentID string `json:"incidentId"`
}
