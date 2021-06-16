#! make

SHELL := /bin/bash

.PHONY: coverage
coverage:
	@echo "Create coverprofile"
	@exec go test -coverprofile=cover.out -v ./pkg/...
	@echo "Generate cover.html"
	@exec go tool cover -html=cover.out -o cover.html

.PHONY: swagger
swagger:
	@echo "Generate Swagger documentation"
	@exec ~/go/bin/swag init -g pkg/api/api.go