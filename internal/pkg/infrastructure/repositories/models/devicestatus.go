package models

type DeviceStatus struct {
	DeviceId string
	Status   string
}

type StatusMessage struct {
	DeviceID     string   `json:"deviceID"`
	BatteryLevel int      `json:"batteryLevel"`
	Status       int      `json:"statusCode"`
	Messages     []string `json:"statusMessages"`
	Timestamp    string   `json:"timestamp"`
}

func NewStatusMessage(deviceID string, code int) StatusMessage {
	return StatusMessage{
		DeviceID: deviceID,
		Status: code,
	}
}
