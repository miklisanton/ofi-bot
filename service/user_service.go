package service

import (
	"ofibot/repositories"
)

type UserService struct {
	repo *repositories.UserRepo
}

func NewUserService(repo *repositories.UserRepo) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) AddUser(chatID int64) error {
	return s.repo.InsertUser(chatID)
}
