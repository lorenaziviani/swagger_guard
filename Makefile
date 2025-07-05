APP=swagger_guard

.PHONY: all test build run lint docker docker-run metrics metrics-local security

all: build

build:
	docker-compose build

test:
	go test -v ./...

lint:
	golangci-lint run || true

security:
	gosec ./...

docker-up:
	docker-compose up -d

docker-run-dev:
	docker-compose up $(APP)

metrics:
	docker-compose run --rm swagger_guard parse --metrics