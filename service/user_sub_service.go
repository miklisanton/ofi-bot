package service

import (
	"ofibot/repositories"
)

type UserSubService struct {
	repo *repositories.UserSubRepo
}

func NewUserSubService(repo *repositories.UserSubRepo) *UserSubService {
	return &UserSubService{repo: repo}
}
