package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"

	"gitlab.ozon.ru/university/broker-4/config"
	"gitlab.ozon.ru/university/broker-4/pkg/rabbitmq"
	"gitlab.ozon.ru/university/broker-4/pkg/sqlite"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "reset-db:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	queues := flag.Bool("queues", false, "also clear application queues; stop the server and clients first")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("usage: reset-db [-queues]")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), config.OperationTimeout)
	defer cancel()
	if *queues {
		if err = resetQueues(ctx, cfg.RabbitURL); err != nil {
			return err
		}
		fmt.Println("api.requests, api.retry and api.dlq cleared; broker users and configuration preserved")
	}
	db, err := sqlite.Open(ctx, cfg.Database)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		rollbackErr := tx.Rollback()
		if !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, rollbackErr)
		}
	}()
	_, err = tx.ExecContext(ctx, `DELETE FROM processed_requests; DELETE FROM tasks; DELETE FROM projects; DELETE FROM sqlite_sequence WHERE name IN ('projects', 'tasks');`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	fmt.Println("projects, tasks and processed requests cleared; IDs reset")
	return nil
}

func resetQueues(ctx context.Context, address string) (err error) {
	session, err := rabbitmq.Open(ctx, address)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, session.Close()) }()

	err = session.ResetQueues(ctx)
	return err
}
