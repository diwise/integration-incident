package models

import (
	"strconv"
	"time"
)

type DeviceStatus struct {
	DeviceId string
	Status   string
}

type StatusMessage struct {
	DeviceID string `json:"deviceID"`

	BatteryLevel *float64 `json:"batteryLevel,omitempty"`

	Code     *string  `json:"statusCode,omitempty"`
	Messages []string `json:"statusMessages,omitempty"`

	RSSI            *float64 `json:"rssi,omitempty"`
	LoRaSNR         *float64 `json:"loRaSNR,omitempty"`
	Frequency       *int64   `json:"frequency,omitempty"`
	SpreadingFactor *float64 `json:"spreadingFactor,omitempty"`
	DR              *int     `json:"dr,omitempty"`

	Tenant    string    `json:"tenant"`
	Timestamp time.Time `json:"timestamp"`
}

func NewStatusMessage(deviceID string, code int) StatusMessage {
	c := strconv.Itoa(code)

	return StatusMessage{
		DeviceID: deviceID,
		Code:     &c,
	}
}
