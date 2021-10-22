# Billing system test task

[![Go](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/vsokoltsov/billing_system_test_task_go/actions/workflows/go.yml)

## Code coverage

```shell
billing_system_test_task/internal/pipeline/pipeline.go:11:				ExecutePipeline			100.0%
billing_system_test_task/internal/pipeline/pipeline.go:24:				executeJob			100.0%
billing_system_test_task/internal/repositories/operation.go:35:			NewWalletOperationRepo		100.0%
billing_system_test_task/internal/repositories/operation.go:41:			WithTx				100.0%
billing_system_test_task/internal/repositories/operation.go:45:			Create				100.0%
billing_system_test_task/internal/repositories/operation.go:77:			List				100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:41:		NewFileStorage			100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:46:		NewFileHandler			100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:53:		Create				100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:58:		Create				100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:87:		CreateMarshaller		100.0%
billing_system_test_task/internal/repositories/reports/file_handler.go:116:		GetFileMetadata			100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:30:	NewJSONHandler			100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:39:	MarshallOperation		100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:53:	WriteToFile			100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:71:	NewCSVHandler			100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:79:	MarshallOperation		100.0%
billing_system_test_task/internal/repositories/reports/file_marshaller.go:100:	WriteToFile			100.0%
billing_system_test_task/internal/repositories/reports/pipes.go:20:			NewOperationsProcessesManager	100.0%
billing_system_test_task/internal/repositories/reports/pipes.go:25:			Process				100.0%
billing_system_test_task/internal/repositories/reports/pipes.go:75:			Call				100.0%
billing_system_test_task/internal/repositories/reports/pipes.go:102:			Call				100.0%
billing_system_test_task/internal/repositories/reports/pipes.go:130:			Call				100.0%
billing_system_test_task/internal/repositories/reports/query_params.go:24:		NewQueryParamsReader		100.0%
billing_system_test_task/internal/repositories/reports/query_params.go:29:		Parse				100.0%
billing_system_test_task/internal/repositories/users.go:30:				NewUsersService			100.0%
billing_system_test_task/internal/repositories/users.go:36:				WithTx				100.0%
billing_system_test_task/internal/repositories/users.go:41:				GetByID				100.0%
billing_system_test_task/internal/repositories/users.go:75:				GetByWalletID			100.0%
billing_system_test_task/internal/repositories/users.go:108:				Create				100.0%
billing_system_test_task/internal/repositories/wallet.go:34:				NewWalletService		100.0%
billing_system_test_task/internal/repositories/wallet.go:40:				WithTx				100.0%
billing_system_test_task/internal/repositories/wallet.go:45:				Create				100.0%
billing_system_test_task/internal/repositories/wallet.go:76:				Enroll				100.0%
billing_system_test_task/internal/repositories/wallet.go:132:			GetByID				100.0%
billing_system_test_task/internal/repositories/wallet.go:144:			GetByUserId			100.0%
billing_system_test_task/internal/repositories/wallet.go:156:			Transfer			100.0%
billing_system_test_task/internal/transport/http/api.go:21:				NewRouter			100.0%
billing_system_test_task/internal/transport/http/operation.go:17:			NewOperationsHandler		100.0%
billing_system_test_task/internal/transport/http/operation.go:36:			List				100.0%
billing_system_test_task/internal/transport/http/user.go:27:				NewUserHandler			100.0%
billing_system_test_task/internal/transport/http/user.go:44:				Create				100.0%
billing_system_test_task/internal/transport/http/user.go:92:				Enroll				100.0%
billing_system_test_task/internal/transport/http/wallet.go:21:			NewWalletsHandler		100.0%
billing_system_test_task/internal/transport/http/wallet.go:38:			Transfer			100.0%
billing_system_test_task/internal/usecases/operation.go:24:				NewWalletOperationInteractor	100.0%
billing_system_test_task/internal/usecases/operation.go:34:				GenerateReport			100.0%
billing_system_test_task/internal/usecases/user.go:27:				NewUserInteractor		100.0%
billing_system_test_task/internal/usecases/user.go:38:				Create				100.0%
billing_system_test_task/internal/usecases/user.go:78:				Enroll				100.0%
billing_system_test_task/internal/usecases/wallet.go:24:				NewWalletInteractor		100.0%
billing_system_test_task/internal/usecases/wallet.go:33:				Transfer			100.0%
total:										(statements)			100.0%
```

## How to run

1. Populate `.env` file with necessary variables (you can find them in `.env.sample`)
2. Run `docker compose up`, which runs application in development environment

* If you need to down all migrations, enter in the app container and run `make migrations-down`

## Test

* For testing use `make test`
* `make coverage` generates cover files (which are already present)

## Lint

* `make lint` validates code agains [gofmt](https://internal.go.dev/cmd/gofmt) and [golangci-lint](https://github.com/golangci/golangci-lint)

## Documentation

* Information about endpoints stored in Swagger documentation, which is available on `/swagger/index.html` endpoint

## Benchmarking

* For generating benchmark files run `make benchmark package=<package>`, where 
    * `<package>` - name of the package for which benchmarks should be generated
* For displaying benchmarks in web ui run `make benchmark-ui package=<package> param=<param>`, where 
  * `<package>` - name of the package for which benchmarks should be generated
  * `<param>` - name of the parameter, that was generated on `make benchmark` step (`cpu` or `mem`)