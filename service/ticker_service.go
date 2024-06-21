package service

import (
	"ofibot/models"
	"ofibot/repositories"
)

type TickerService struct {
	repo *repositories.TickerRepo
}

func NewTickerService(repo *repositories.TickerRepo) *TickerService {
	return &TickerService{repo: repo}
}

func (s *TickerService) GetAllTickers() ([]models.Ticker, error) {
	return s.repo.GetAllTickers()
}

func (s *TickerService) GetSubscribedTickers() ([]string, error) {
	return s.repo.GetSubscribedTickers()
}
