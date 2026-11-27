.PHONY: api run build lint fmt

GO ?= $(if $(GOROOT),$(GOROOT)/bin/go,go)
PYTHON ?= python3
API_PORT ?= 8000

api: .venv/.requirements-installed
	.venv/bin/python -B -m uvicorn main:app --app-dir API --host 127.0.0.1 --port "$(API_PORT)"

.venv/bin/python:
	uv venv --python "$(PYTHON)" .venv

.venv/.requirements-installed: API/requirements.txt | .venv/bin/python
	uv pip install --python .venv/bin/python -r API/requirements.txt
	@touch $@

run:
	"$(GO)" run ./cmd/client

build:
	"$(GO)" build -o bin/client ./cmd/client

lint:
	golangci-lint run ./...

fmt:
	goimports -w cmd internal pkg
