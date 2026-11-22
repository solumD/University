package app

import (
	"context"
	"errors"
	"syscall"

	"gitlab.ozon.ru/university/broker-4/config"
	delivery "gitlab.ozon.ru/university/broker-4/internal/delivery/amqp"
	repository "gitlab.ozon.ru/university/broker-4/internal/repository/sqlite"
	"gitlab.ozon.ru/university/broker-4/internal/usecase"
	"gitlab.ozon.ru/university/broker-4/pkg/logger"
	"gitlab.ozon.ru/university/broker-4/pkg/rabbitmq"
	"gitlab.ozon.ru/university/broker-4/pkg/sqlite"
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
		syncErr := log.Sync()
		if !errors.Is(syncErr, syscall.EINVAL) && !errors.Is(syncErr, syscall.ENOTTY) {
			err = errors.Join(err, syncErr)
		}
	}()

	db, err := sqlite.Open(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	session, err := rabbitmq.Open(ctx, cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, session.Close()) }()

	storage := repository.New(db)
	executor := usecase.NewExecutor(storage)
	server := delivery.New(session, executor, cfg, log)
	err = server.Run(ctx)
	return err
}
