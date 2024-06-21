package repositories

import (
	"database/sql"
	"fmt"
	"ofibot/models"
)

type TickerRepo struct {
	db *sql.DB
}

func NewTickerRepo(db *sql.DB) *TickerRepo {
	return &TickerRepo{db: db}
}

func (repo *TickerRepo) GetAllTickers() ([]models.Ticker, error) {
	rows, err := repo.db.Query("SELECT rank, symbol FROM ticker")
	if err != nil {
		return nil, fmt.Errorf("Failed to get all tickers: ", err)
	}
	defer rows.Close()

	var tickers []models.Ticker
	for rows.Next() {
		var ticker models.Ticker
		err := rows.Scan(&ticker.ID, &ticker.Symbol)
		if err != nil {
			return nil, fmt.Errorf("Failed to get all tickers: ", err)
		}
		tickers = append(tickers, ticker)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to get all tickers: ", err)
	}

	return tickers, nil
}

func (repo *TickerRepo) GetSubscribedTickers() ([]string, error) {
	query := "SELECT DISTINCT symbol FROM subscription s JOIN ticker t ON s.ticker_id = t.symbol"

	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Failed to get subscribed tickers: ", err)
	}
	defer rows.Close()

	var tickers []string
	for rows.Next() {
		var ticker string
		err := rows.Scan(&ticker)
		if err != nil {
			return nil, fmt.Errorf("Failed to get subscribed tickers: ", err)
		}
		tickers = append(tickers, ticker)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Failed to get subscribed tickers: ", err)
	}

	return tickers, nil
}
