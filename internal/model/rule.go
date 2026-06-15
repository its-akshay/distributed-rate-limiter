package model

import "time"

type Rule struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	LimitCount    int       `json:"limit_count"`
	WindowSeconds int       `json:"window_seconds"`
	CreatedAt     time.Time `json:"created_at"`
}