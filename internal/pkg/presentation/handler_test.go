package presentation

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestNotificationHandler(t *testing.T) {
	is := is.New(t)

	server := setupMockService(http.StatusOK, "")

	r := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(notificationJson)))
	w := httptest.NewRecorder()

	notificationHandler().ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)
}

func setupMockService(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/ld+json")
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}))
}

const notificationJson string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:05393925","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41}}]
}`
