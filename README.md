# Billing system test task

[![Go](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml)

## How to run

1. Populate `.env` file with necessary variables (you can find them in `.env.sample`)
2. Run `docker compose up`, which runs application in development environment

* If you need to down all migrations, enter in the app container and run `make migrations-down`

## Test

* For testing use `make test`
* `make coverage` generates cover files (which are already present)

## Lint

* `make lint` validates code agains [gofmt](https://pkg.go.dev/cmd/gofmt) and [golangci-lint](https://github.com/golangci/golangci-lint)

## Documentation

* Information about endpoints stored in Swagger documentation, which is available on `/swagger/index.html` endpoint

## Benchmarking

* For generating benchmark files run `make benchmark package=<package>`, where 
    * `<package>` - name of the package for which benchmarks should be generated
* For displaying benchmarks in web ui run `make benchmark-ui package=<package> param=<param>`, where 
  * `<package>` - name of the package for which benchmarks should be generated
  * `<param>` - name of the parameter, that was generated on `make benchmark` step (`cpu` or `mem`)