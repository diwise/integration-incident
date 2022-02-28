package application

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestNotificationHandler(t *testing.T) {
	is := is.New(t)

	server := setupMockService([]response{
		{http.StatusOK, ""},
	})

	r := httptest.NewRequest("POST", server.URL+"/notification", bytes.NewBuffer([]byte(notificationJson)))
	w := httptest.NewRecorder()

	notificationHandler().ServeHTTP(w, r)
	is.Equal(w.Code, http.StatusOK)
}

const notificationJson string = `{"subscriptionId":"36990e41ccd84af99d8b233eca81d1d3","data":[{"id":"urn:ngsi-ld:Device:se:servanet:lora:msva:05393925","type":"Device","rssi":{"type":"Property","value":0.1},"snr":{"type":"Property","value":0.41}}]
}`
