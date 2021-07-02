package models

type Incident struct {
	DeviceId    string     `json:"deviceId"`
	Category    int        `json:"category"`
	Description string     `json:"description"`
	Coordinates [2]float64 `json:"coordinates"`
}
