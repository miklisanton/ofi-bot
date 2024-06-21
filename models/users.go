package models

import "time"

type User struct {
	ID      int
	ChatID  int64
	Created time.Time
}
