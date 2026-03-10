# Load .env if exists
ifneq (,$(wildcard .env))
include .env
export
endif

APP=btree-cli
BIN=bin

.PHONY: default tidy

default: ## show all available tasks
	@make help

tidy: ## install pkg listed in go.mod
	go mod tidy

help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

governance:
	go mod tidy -v
	go fmt ./...
	go vet ./...

build:
	go build -o $(BIN)/$(APP) .

clean:
	rm -rf $(BIN)
