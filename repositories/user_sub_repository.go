package repositories

import (
	"database/sql"
)

type UserSubRepo struct {
	db *sql.DB
}

func NewUserSubRepo(db *sql.DB) *UserSubRepo {
	return &UserSubRepo{db: db}
}
