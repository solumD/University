package app

import (
	"context"
	"errors"
	"syscall"

	"gitlab.ozon.ru/university/api-2/config"
	delivery "gitlab.ozon.ru/university/api-2/internal/delivery/http"
	repository "gitlab.ozon.ru/university/api-2/internal/repository/sqlite"
	"gitlab.ozon.ru/university/api-2/internal/usecase"
	httpserver "gitlab.ozon.ru/university/api-2/pkg/http_server"
	"gitlab.ozon.ru/university/api-2/pkg/logger"
	"gitlab.ozon.ru/university/api-2/pkg/sqlite"
	"go.uber.org/zap"
)

func Run(ctx context.Context) (err error) {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return err
	}

	defer func() {
		if syncErr := log.Sync(); syncErr != nil && !errors.Is(syncErr, syscall.EINVAL) && !errors.Is(syncErr, syscall.ENOTTY) {
			err = errors.Join(err, syncErr)
		}
	}()

	db, err := sqlite.Open(ctx, cfg.Database)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := db.Close()
		err = errors.Join(err, closeErr)
	}()

	service := usecase.New(repository.New(db))
	router := delivery.NewRouter(cfg, service, log)
	server := httpserver.New(cfg.Address, router, log)
	log.Info("starting server", zap.String("address", cfg.Address))
	err = server.Run(ctx)

	return err
}
