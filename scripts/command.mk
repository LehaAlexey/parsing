.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build ./...

