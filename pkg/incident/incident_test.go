package incident

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diwise/integration-incident/internal/pkg/infrastructure/repositories/models"
)

func TestPostLifebuouyIncident(t *testing.T) {

	server := setupMockService(http.StatusOK, accessTokenResp)

	incidentReporter, _ := NewIncidentReporter(context.Background(), server.URL, "")

	incident := models.Incident{
		PersonId:       "deviceID",
		Description:    "description",
		Category:       5,
		MapCoordinates: "62.0,17.0",
	}

	err := incidentReporter(context.Background(), incident)
	if err != nil {
		t.Errorf("could not post incident: %s", err.Error())
	}
}

func setupMockService(responseCode int, responseBody string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "token") {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(accessTokenResp))
		} else {
			w.Header().Add("Content-Type", "application/ld+json")
			w.WriteHeader(responseCode)
			w.Write([]byte(`{
				"status": "INSKICKAT",
				"incidentId": "SP_20210819_415b"
			  }`))
		}
	}))
}

const accessTokenResp string = `{"access_token":"ncjklhclabclksabclac",
"scope":"am_application_scope default",
"token_type":"Bearer",
"expires_in":3600}
`
