package models

type UserSubscription struct {
	ChatID   int64
	UserID   int
	TickerID int
	Symbol   string
}
