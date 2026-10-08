package models

import "time"

type Job struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	URL       string    `json:"url" bson:"url"`
	Width     int       `json:"width" bson:"width"`
	Height    int       `json:"height" bson:"height"`
	Status    string    `json:"status" bson:"status"` // ex. "pending", "processing", "completed", "failed"
	ResultURL string    `json:"result_url,omitempty" bson:"result_url,omitempty"`
	Error     string    `json:"error,omitempty" bson:"error,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
