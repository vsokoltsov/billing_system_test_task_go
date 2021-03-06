#! make

SHELL := /bin/bash

.PHONY: coverage
coverage:
	@echo "Create coverprofile"
	@exec go test -coverprofile=cover.out.tmp -v ./internal/...
	@exec cat cover.out.tmp | grep -v "_mock.go" > cover.out
	@echo "Generate cover.html"
	@exec go tool cover -html=cover.out -o cover.html

.PHONY: coverage-all
coverage-all:
	@echo "Create coverprofile"
	@exec go test -coverprofile=cover.out.tmp -v ./internal/...
	@echo "Show full coverage"
	@exec go tool cover  -func=cover.out

.PHONY: swagger
swagger:
	@echo "Generate Swagger documentation"
	@exec ~/go/bin/swag init -g internal/transport/http/api.go

.PHONY: build
build:
	@echo "Build application server"
	@exec go build -o ./tmp/app/server cmd/billing/main.go

.PHONY: run-server
run-server:
	make migrations-up
	@echo "Run server application"
	@exec air

.PHONY: migrations-up
migrations-up:
	@echo "Run migrations up"
	@exec migrate -path ./migrations -database ${DB_CON} up

.PHONY: migrations-down
migrations-down:
	@echo "Run migrations down"
	@exec migrate -path ./migrations -database ${DB_CON} down -all


.PHONY: test
test:
	@echo "Run tests (without coverage)"
	@exec go test -v ./internal/...

.PHONY: lint
lint:
	@echo "Check via gofmt..."
	@exec gofmt -w ./
	@echo "Check via golangci-lint..."
	@exec golangci-lint run

.PHONY: benchmark
benchmark:
	@echo "Run benchmark for '$(package)' package"
	@exec go test -v ./internal/$(package)/... -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1

.PHONY: benchmark-ui
benchmark-ui:
	make benchmark package=$(package)
	@exec go tool pprof -http=:8091 $(package).test $(param).out