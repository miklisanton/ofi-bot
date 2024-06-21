package service

import (
	"fmt"
	"ofibot/models"
	"ofibot/repositories"
)

type SubscriptionService struct {
	repo *repositories.SubscriptionRepo
}

func NewSubscriptionService(repo *repositories.SubscriptionRepo) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) AddSubscription(chatID int64, symbol string) (*models.Subscription, error) {
	return s.repo.InsertSubscription(chatID, symbol)
}

func (s *SubscriptionService) RemoveSubscription(chatID int64, symbol string) error {
	return s.repo.RemoveSubscription(chatID, symbol)
}

func (s *SubscriptionService) GetSubscriptionsString(tickers []models.Ticker) (string, error) {
	if len(tickers) == 0 {
		return "", fmt.Errorf("no tickers")
	}

	out := "You are subscribed to tickers:\n"
	for _, subscription := range tickers {
		out += subscription.Symbol + ", "
	}
	out = out[:len(out)-2]
	return out, nil
}

func (s *SubscriptionService) GetUserTickers(chatID int64) ([]models.Ticker, error) {
	subscriptions, err := s.repo.GetUserSubscriptions(chatID)
	if err != nil {
		return nil, err
	}

	var tickers []models.Ticker
	for _, subscription := range subscriptions {
		tickers = append(tickers, models.Ticker{Symbol: subscription.TickerID, ID: -1})
	}
	return tickers, nil
}

func (s *SubscriptionService) GetUsersSubscribed(symbol string) ([]int64, error) {
	return s.repo.GetUsersSubscribed(symbol)
}
