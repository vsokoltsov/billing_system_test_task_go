# Billing system test task

## How to run

1. Populate `.env` file with necessary variables (you can find them in `.env.sample`)
2. run `make up`, which runs application in development environment

## Test

* For testing use `make test`
* `make coverage` generates cover files (which are already present)

## Lint

* `make lint` validates code agains [gofmt](https://pkg.go.dev/cmd/gofmt) and [golangci-lint](https://github.com/golangci/golangci-lint)

## Documentation

* Information about endpoints stored in Swagger documentation, which is available on `/swagger/index.html` endpoint