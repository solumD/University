package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gitlab.ozon.ru/university/api-1/migrations"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, path string) (*sql.DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(filepath.Dir(abs), 0700); err != nil {
		return nil, err
	}

	uri := url.URL{Scheme: "file", Path: abs}
	query := url.Values{"_pragma": {"busy_timeout(5000)", "journal_mode(WAL)"}}
	uri.RawQuery = query.Encode()
	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		return nil, err
	}

	if err = migrate(ctx, db); err != nil {
		closeErr := db.Close()
		err = errors.Join(err, closeErr)

		return nil, err
	}

	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var err error

	if _, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}

	files, err := migrations.Files.ReadDir(".")
	if err != nil {
		return err
	}

	for _, file := range files {
		var count int

		if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, file.Name()).Scan(&count); err != nil {
			return err
		}

		if count != 0 {
			continue
		}

		body, readErr := migrations.Files.ReadFile(file.Name())
		if readErr != nil {
			return readErr
		}

		if _, err = db.ExecContext(ctx, string(body)); err != nil {
			err = fmt.Errorf("apply migration %s: %w", file.Name(), err)

			return err
		}

		if _, err = db.ExecContext(ctx, `INSERT INTO schema_migrations (name) VALUES (?)`, file.Name()); err != nil {
			return err
		}
	}

	return nil
}
