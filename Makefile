#! make

SHELL := /bin/bash

.PHONY: coverage
coverage:
	@echo "Create coverprofile"
	@exec go test -coverprofile=cover.out.tmp -v ./pkg/...
	@exec cat cover.out.tmp | grep -v "_mock.go" > cover.out
	@echo "Generate cover.html"
	@exec go tool cover -html=cover.out -o cover.html

.PHONY: swagger
swagger:
	@echo "Generate Swagger documentation"
	@exec ~/go/bin/swag init -g pkg/api/api.go

.PHONY: build
build:
	@echo "Build application server"
	@exec go build -o ./tmp/app/server cmd/billing/main.go

.PHONY: up
up:
	@echo "Up docker-compose project"
	@exec docker compose up

.PHONY: test
test:
	@echo "Run tests (without coverage)"
	@exec go test -v ./pkg/...

.PHONY: lint
lint:
	@echo "Check via gofmt..."
	@exec gofmt -w ./
	@echo "Check via golangci-lint..."
	@exec golangci-lint run