package api

type Notification struct {
	Id             string `json:"id"`
	Type           string `json:"type"`
	SubscriptionId string `json:"subscriptionId"`
	NotifiedAt     string `json:"notifiedAt"`
	Data           []struct {
		Id   string `json:"id"`
		Type string `json:"type"`
		Status *struct {
			Value string `json:"value"`
		} `json:"status,omitempty"`
		DeviceState *struct {
			Value string `json:"value"`
		} `json:"deviceState,omitempty"`				
	} `json:"data"`
}
