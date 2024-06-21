package models

import "time"

type Subscription struct {
	ID       int
	TickerID string
	UserID   int
	Created  time.Time
}
