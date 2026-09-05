.PHONY: build run test lint migrate-up migrate-down migrate-revert docker-up docker-down tidy templ-gen clean

BIN=bin/ezauth-example

build: templ-gen
	go build -o $(BIN) ./cmd/server

run: build
	./$(BIN)

test:
	go test ./... -count=1

lint:
	go vet ./...
	golangci-lint run

migrate-up: build
	./$(BIN) -migrate up

migrate-down: build
	./$(BIN) -migrate down

migrate-revert: build
	./$(BIN) -migrate revert

docker-up:
	docker compose up --build

docker-down:
	docker compose down

templ-gen:
	templ generate

tidy:
	go mod tidy

clean:
	rm -rf bin/ views/**/*_templ.go
