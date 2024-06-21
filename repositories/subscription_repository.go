package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"ofibot/models"
)

type SubscriptionRepo struct {
	db *sql.DB
}

func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

func (repo *SubscriptionRepo) InsertSubscription(chatID int64, symbol string) (*models.Subscription, error) {
	query := `
		INSERT INTO subscription (user_id, ticker_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, ticker_id) DO NOTHING
		RETURNING id, user_id, ticker_id, created_at
`
	var subscription models.Subscription

	err := repo.db.QueryRow(query, chatID, symbol).Scan(&subscription.ID,
		&subscription.UserID,
		&subscription.TickerID,
		&subscription.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, err
}

func (repo *SubscriptionRepo) RemoveSubscription(chatID int64, symbol string) error {
	query := `DELETE FROM subscription WHERE user_id = $1 AND ticker_id = $2;`
	_, err := repo.db.Exec(query, chatID, symbol)
	if err != nil {
		return err
	}

	return nil
}

func (repo *SubscriptionRepo) GetUserSubscriptions(chatID int64) ([]models.Subscription, error) {
	query := `SELECT id, user_id, ticker_id, created_at FROM subscription WHERE user_id = $1`

	rows, err := repo.db.Query(query, chatID)

	if err != nil {
		return nil, fmt.Errorf("Failed to get subscriptions for user ", chatID, err)
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	for rows.Next() {
		var subscription models.Subscription
		err := rows.Scan(&subscription.ID, &subscription.UserID, &subscription.TickerID, &subscription.Created)
		if err != nil {
			return nil, fmt.Errorf("Failed to get subscriptions for user ", chatID, err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to get subscriptions for user ", chatID, err)
	}

	return subscriptions, nil
}

func (repo *SubscriptionRepo) GetUsersSubscribed(symbol string) ([]int64, error) {
	query := `SELECT user_id FROM subscription WHERE ticker_id = $1`

	rows, err := repo.db.Query(query, symbol)

	if err != nil {
		return nil, fmt.Errorf("Failed to get users subscribed on ", symbol, err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("Failed to get users subscribed on ", symbol, err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to get users subscribed on ", symbol, err)
	}

	return ids, nil
}
