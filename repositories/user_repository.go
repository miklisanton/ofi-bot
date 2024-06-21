package repositories

import (
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (repo *UserRepo) InsertUser(chatID int64) error {
	_, err := repo.db.Exec("INSERT INTO users (chat_id) VALUES ($1) ON CONFLICT (chat_id) DO NOTHING;", chatID)
	return err
}
