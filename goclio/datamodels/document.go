package datamodels

import "time"

type Document struct {
	Id        int       `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Matter    Matter    `json:"matter"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
