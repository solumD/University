package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gitlab.ozon.ru/university/api-1/internal/model"
)

func (r *Repository) Cached(ctx context.Context, key model.CommandKey, now time.Time) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	if _, err := r.db.ExecContext(ctx, `DELETE FROM idempotency WHERE expires_at <= ?`, now.UTC().Format(time.RFC3339)); err != nil {
		return "", nil, err
	}

	var fingerprint string

	var response []byte

	err := r.db.QueryRowContext(ctx, `SELECT fingerprint, response FROM idempotency WHERE scope = ? AND key = ?`, key.Scope, key.Key).
		Scan(&fingerprint, &response)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, nil
	}

	return fingerprint, response, err
}

func (r *Repository) SaveCached(ctx context.Context, key model.CommandKey, fingerprint string, response []byte, expires time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `INSERT INTO idempotency (scope, key, fingerprint, response, expires_at) VALUES (?, ?, ?, ?, ?)`,
		key.Scope, key.Key, fingerprint, response, expires.UTC().Format(time.RFC3339))

	return err
}
