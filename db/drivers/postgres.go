package postgres

import (
	// import go libpq driver package
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Connect(connectionURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if err = runMigrations(db); err != nil {
		return nil, err
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")

	if err := goose.Up(db, "db/migrations"); err != nil {
		return err
	}
	return nil
}
