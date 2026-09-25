# Go Binary & App Variables
APP_NAME := api-server
MAIN_PATH := ./cmd/api
GO ?= go

.PHONY: all build run test clean tidy

all: test build

## run: Run the development server
run:
	$(GO) run $(MAIN_PATH)

## build: Build the server binary into bin/
build:
	@mkdir -p bin
	$(GO) build -o bin/$(APP_NAME) $(MAIN_PATH)

## test: Run unit tests with verbose output
test:
	$(GO) test -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

## tidy: Tidy go.mod dependencies
tidy:
	$(GO) mod tidy

## clean: Remove build artifacts
clean:
	@rm -rf bin coverage.out
