package api

type Notification struct {
	Id             string                   `json:"id"`
	Type           string                   `json:"type"`
	SubscriptionId string                   `json:"subscriptionId"`
	NotifiedAt     string                   `json:"notifiedAt"`
	Data           []map[string]interface{} `json:"data"`
}
