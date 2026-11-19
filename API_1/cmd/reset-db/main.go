package main

import (
	"context"
	"errors"
	"os"
	"time"

	"gitlab.ozon.ru/university/api-1/config"
	"gitlab.ozon.ru/university/api-1/pkg/sqlite"
	"go.uber.org/zap"
)

func main() {
	log := zap.NewExample()

	err := reset(log)
	if err != nil {
		log.Error("failed to reset database", zap.Error(err))
		os.Exit(1)
	}
}

func reset(log *zap.Logger) (err error) {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlite.Open(ctx, cfg.Database)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := db.Close()
		err = errors.Join(err, closeErr)
	}()

	queries := []string{
		"DELETE FROM idempotency",
		"DELETE FROM tasks",
		"DELETE FROM projects",
		"DELETE FROM sqlite_sequence WHERE name IN ('projects', 'tasks')",
	}

	for _, query := range queries {
		_, err = db.ExecContext(ctx, query)
		if err != nil {
			return err
		}
	}

	log.Info("database cleared", zap.String("path", cfg.Database))

	return nil
}
