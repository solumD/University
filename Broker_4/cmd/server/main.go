package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"gitlab.ozon.ru/university/broker-4/internal/app"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx); err != nil {
		log := zap.NewExample()
		log.Error("server stopped", zap.Error(err))
		os.Exit(1)
	}
}
