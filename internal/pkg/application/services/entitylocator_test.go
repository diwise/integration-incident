package services

import (
	"context"
	"net/http"
	"testing"

	test "github.com/diwise/service-chassis/pkg/test/http"
	"github.com/diwise/service-chassis/pkg/test/http/expects"
	"github.com/diwise/service-chassis/pkg/test/http/response"

	"github.com/matryer/is"
)

func TestMe(t *testing.T) {
	is := is.New(t)

	ms := test.NewMockServiceThat(
		test.Expects(is, expects.RequestPath("/ngsi-ld/v1/entities/someid")),
		test.Returns(
			response.Code(http.StatusOK),
			response.Body([]byte(pointedEntity)),
		),
	)

	locator, err := NewEntityLocator(ms.URL(), "customTenant")
	is.NoErr(err)

	lat, lon, err := locator.Locate(context.Background(), "Something", "someid")
	is.NoErr(err)

	is.Equal(lat, 62.364)
	is.Equal(lon, 17.371627)
}

const pointedEntity string = `{"dateObserved":{"@type":"DateTime","@value":"2022-11-08T08:00:00Z"},"id":"urn:ngsi-ld:WaterQualityObserved:SE0712281000003480:20221108T080000Z","location":{"type":"Point","coordinates":[17.371627,62.364]},"temperature":5.5,"type":"WaterQualityObserved"}`
