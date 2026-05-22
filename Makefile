.DEFAULT_GOAL := build

BINARY_NAME = dota2-bot
BUILD_PATH = cmd/build

.SILENT:

build:
	mkdir -p $(BUILD_PATH)
	CGO_ENABLED=0 go build -o $(BUILD_PATH)/$(BINARY_NAME) cmd/main.go

clean:
	rm -rf $(BUILD_PATH)

lint:
	golangci-lint run

run:
	go run ./cmd

tidy:
	go mod tidy

vet:
	go vet ./...

# Создать новую миграцию: make migrate-new name=add_something
migrate-new:
	migrate create -ext sql -dir migrations -seq $(name)
