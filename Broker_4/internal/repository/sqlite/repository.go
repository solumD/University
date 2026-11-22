package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gitlab.ozon.ru/university/broker-4/internal/usecase"
)

const queryTimeout = 5 * time.Second

type Storage struct{ db *sql.DB }
type Repository struct{ db *sql.Tx }

func New(db *sql.DB) *Storage {
	storage := &Storage{db: db}
	return storage
}

func (s *Storage) Begin(ctx context.Context) (usecase.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	repo := &Repository{db: tx}
	return repo, nil
}

func (r *Repository) Commit() error {
	err := r.db.Commit()
	return err
}

func (r *Repository) Rollback() error {
	err := r.db.Rollback()
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
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
