package models

type Incident struct {
	PersonId       string   `json:"personId"`
	Category       int      `json:"category"`
	Description    string   `json:"description"`
	MapCoordinates string   `json:"mapCoordinates"`
	Attachments    []string `json:"attachments"`
}
