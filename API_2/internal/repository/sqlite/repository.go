package sqlite

import (
	"database/sql"
	"errors"
	"time"
)

const queryTimeout = 5 * time.Second

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	repo := &Repository{db: db}

	return repo
}

func affected(result sql.Result, err, missing error) error {
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return missing
	}

	return nil
}

func notFound(err, missing error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return missing
	}

	return err
}
