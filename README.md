# Billing system test task

[![Go](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml)

## How to run

1. Populate `.env` file with necessary variables (you can find them in `.env.sample`)
2. Run `make up`, which runs application in development environment
3. Go to application container (`docker compose exec app bash`) and run migrations ([migrate](https://github.com/golang-migrate/migrate) is used for that)

## Test

* For testing use `make test`
* `make coverage` generates cover files (which are already present)

## Lint

* `make lint` validates code agains [gofmt](https://pkg.go.dev/cmd/gofmt) and [golangci-lint](https://github.com/golangci/golangci-lint)

## Documentation

* Information about endpoints stored in Swagger documentation, which is available on `/swagger/index.html` endpoint