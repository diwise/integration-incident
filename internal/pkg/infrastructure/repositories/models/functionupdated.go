package models

import "time"

type location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type FunctionUpdated struct {
	Id       string    `json:"id"`
	Type     string    `json:"type"`
	SubType  string    `json:"subType"`
	Location *location `json:"location,omitempty"`
	Name     string    `json:"name"`
	
	Counter  *struct {
		Counter int  `json:"counter"`
		State   bool `json:"state"`
	} `json:"counter,omitempty"`
	Level *struct {
		Current float64  `json:"current"`
		Percent *float64 `json:"percent,omitempty"`
		Offset  *float64 `json:"offset,omitempty"`
	} `json:"level,omitempty"`
	Presence *struct {
		State bool `json:"state"`
	} `json:"presence,omitempty"`
	Timer *struct {
		StartTime time.Time      `json:"startTime"`
		EndTime   *time.Time     `json:"endTime,omitempty"`
		Duration  *time.Duration `json:"duration,omitempty"`
		State     bool           `json:"state"`
	} `json:"timer,omitempty"`
	WaterQuality *struct {
		Temperature float64   `json:"temperature"`
		Timestamp   time.Time `json:"timestamp"`
	} `json:"waterquality,omitempty"`
	Building *struct {
		Energy float64 `json:"energy"`
		Power  float64 `json:"power"`
	} `json:"building,omitempty"`
	Stopwatch *struct {
		StartTime      time.Time      `json:"startTime"`
		StopTime       *time.Time     `json:"stopTime,omitempty"`
		Duration       *time.Duration `json:"duration,omitempty"`
		State          bool           `json:"state"`
		Count          int32          `json:"count"`
		CumulativeTime time.Duration  `json:"cumulativeTime"`
	}
}
