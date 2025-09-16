package models

import "time"

type Event struct {
	ID     string    `json:"id"`
	UserID int64     `json:"user_id"`
	Date   time.Time `json:"date"`
	Text   string    `json:"event"`
}

type EventsByDay struct {
	UserID int64
	Day    time.Time
}
