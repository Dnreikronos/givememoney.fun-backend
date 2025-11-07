.PHONY: build run test
build:
	@go build -o givememoney.fun-backend cmd/main.go

docker-up:
	@docker-compose up -d

run:
	@go run cmd/main.go

test:
	@go test ./...
