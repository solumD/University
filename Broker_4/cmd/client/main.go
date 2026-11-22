package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"gitlab.ozon.ru/university/broker-4/config"
	"gitlab.ozon.ru/university/broker-4/internal/app"
	"gitlab.ozon.ru/university/broker-4/internal/model"
	"gitlab.ozon.ru/university/broker-4/internal/protocol"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "client:", err)
		os.Exit(1)
	}
}

func run() error {
	version := flag.String("version", "v1", "API version: v1 or v2")
	action := flag.String("action", "", "operation, e.g. create_project or list_tasks")
	data := flag.String("data", "{}", "operation data as JSON")
	id := flag.String("id", "", "request UUID; reuse to repeat the same operation")
	timeout := flag.Duration("timeout", 30*time.Second, "response timeout, e.g. 30s")
	flag.Parse()
	if flag.NArg() != 0 || *action == "" || *timeout <= 0 || !json.Valid([]byte(*data)) {
		validationErr := errors.New("set -action, valid JSON in -data and a positive -timeout; positional arguments are not supported")

		return validationErr
	}

	if *id == "" {
		generated, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		*id = generated.String()
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	request := protocol.Request{Command: model.Command{ID: *id, Version: *version, Action: *action, Data: json.RawMessage(*data)}, Auth: cfg.APIKey}
	err = app.Request(ctx, cfg, request, *timeout, os.Stdout, os.Stderr)
	return err
}
