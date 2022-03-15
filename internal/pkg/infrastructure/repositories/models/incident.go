package models

import "fmt"

type Incident struct {
	PersonId       string   `json:"personId"`
	Category       int      `json:"category"`
	Description    string   `json:"description"`
	MapCoordinates string   `json:"mapCoordinates"`
	Attachments    []string `json:"attachments"`
}

func NewIncident(category int, description string) *Incident {
	return &Incident{
		PersonId:    "diwise",
		Category:    category,
		Description: description,
	}
}

func (i *Incident) AtLocation(latitude, longitude float64) *Incident {
	i.MapCoordinates = fmt.Sprintf("%f,%f", latitude, longitude)
	return i
}
