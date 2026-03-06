# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: default tidy deps

default: ## show all available tasks
	@make help

tidy: ## install pkg listed in go.mod
	go mod tidy

deps: ## install dependencies
	go install github.com/vektra/mockery/v3@v3.4.0
	go install github.com/wadey/gocovmerge@latest

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

governance:
	go mod tidy -v
	go fmt ./...
	go vet ./...