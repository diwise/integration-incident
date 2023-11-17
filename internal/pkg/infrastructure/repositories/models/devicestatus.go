package models

type DeviceStatus struct {
	DeviceId string
	Status   string
}

type StatusMessage struct {
	DeviceID     string   `json:"deviceID"`
	BatteryLevel int      `json:"batteryLevel"`
	Code         int      `json:"statusCode"`
	Messages     []string `json:"statusMessages,omitempty"`
	Tenant       string   `json:"tenant"`
	Timestamp    string   `json:"timestamp"`
}

func NewStatusMessage(deviceID string, code int) StatusMessage {
	return StatusMessage{
		DeviceID: deviceID,
		Code:     code,
	}
}
