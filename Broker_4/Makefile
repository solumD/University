.PHONY: setup broker-up broker-down queues run build lint fmt reset-db reset-all
GO ?= $(if $(GOROOT),$(GOROOT)/bin/go,go)

setup:
	"$(GO)" run ./cmd/setup
broker-up:
	"$(GO)" run ./cmd/broker up
broker-down:
	"$(GO)" run ./cmd/broker down
queues:
	"$(GO)" run ./cmd/broker queues
run:
	"$(GO)" run ./cmd/server
build:
	"$(GO)" build -o bin/server ./cmd/server
	"$(GO)" build -o bin/client ./cmd/client
lint:
	golangci-lint run ./...
fmt:
	goimports -w cmd config internal pkg migrations
reset-db:
	"$(GO)" run ./cmd/reset-db
reset-all:
	"$(GO)" run ./cmd/reset-db -queues
