.PHONY: run reset-db build lint fmt

# Use the Go installation selected by GOROOT when it is explicitly set.
GO ?= $(if $(GOROOT),$(GOROOT)/bin/go,go)

run:
	"$(GO)" run ./cmd/app

reset-db:
	"$(GO)" run ./cmd/reset-db

build:
	$(GO) build -o bin/api ./cmd/app

lint:
	golangci-lint run ./...

fmt:
	goimports -w cmd config internal pkg migrations docs
