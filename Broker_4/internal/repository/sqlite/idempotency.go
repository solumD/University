package sqlite

import (
	"context"
	"database/sql"
	"errors"
)

func (r *Repository) Cached(ctx context.Context, scope, id string) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var fingerprint string
	var response []byte
	err := r.db.QueryRowContext(ctx, `SELECT fingerprint, response FROM processed_requests WHERE scope = ? AND id = ?`, scope, id).Scan(&fingerprint, &response)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, nil
	}
	return fingerprint, response, err
}

func (r *Repository) SaveCached(ctx context.Context, scope, id, fingerprint string, response []byte) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `INSERT INTO processed_requests (scope, id, fingerprint, response) VALUES (?, ?, ?, ?)`, scope, id, fingerprint, response)
	return err
}
