package models

type DeviceStatus struct {
	DeviceId string
	Status   string
}

type StatusMessage struct {
	DeviceID  string  `json:"deviceID"`
	Error     *string `json:"error,omitempty"`
	Status    Status  `json:"status"`
	Timestamp string  `json:"timestamp"`
}

type Status struct {
	Code     int      `json:"statusCode"`
	Messages []string `json:"statusMessages,omitempty"`
}

func NewStatusMessage(deviceID string, code int) StatusMessage {
	return StatusMessage{
		DeviceID: deviceID,
		Status: Status{
			Code: code,
		},
	}
}
