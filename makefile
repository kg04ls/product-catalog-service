.PHONY: migrate test run

migrate:
	docker-compose up -d spanner && docker-compose run --rm spanner-init

test:
	go test ./...

run:
	go run ./cmd/server
