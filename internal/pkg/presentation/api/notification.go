package api

import "github.com/diwise/ngsi-ld-golang/pkg/datamodels/fiware"

type Notification struct {
	SubscriptionID string          `json:"subscriptionId"`
	Data           []fiware.Device `json:"data"`
}
