.PHONY: build run stop clean logs shell help push tag test lint tidy fmt go-build install ci

IMAGE_NAME := tailswan
CONTAINER_NAME := tailswan
REGISTRY := docker-registry.home.flixen.se
TAG := latest

help:
	@echo "TailSwan Makefile Commands"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make build       - Build the Docker image"
	@echo "  make push        - Build, tag and push image to registry"
	@echo "  make tag         - Tag the current image for registry"
	@echo "  make run         - Run the container with docker-compose"
	@echo "  make stop        - Stop the container"
	@echo "  make clean       - Stop and remove container and volumes"
	@echo "  make logs        - Show container logs"
	@echo "  make shell       - Open a shell in the running container"
	@echo "  make status      - Show TailSwan status"
	@echo "  make rebuild     - Rebuild and restart the container"
	@echo ""
	@echo "Go Development Commands:"
	@echo "  make go-build    - Build the Go binary"
	@echo "  make test        - Run Go tests"
	@echo "  make lint        - Run golangci-lint"
	@echo "  make tidy        - Run go mod tidy and verify"
	@echo "  make fmt         - Format Go code"
	@echo "  make install     - Install the binary"
	@echo "  make ci          - Run all CI checks (tidy, lint, test)"
	@echo ""

build:
	@echo "Building TailSwan Docker image..."
	docker build -t $(IMAGE_NAME):latest .

run:
	@echo "Starting TailSwan..."
	docker compose up -d

stop:
	@echo "Stopping TailSwan..."
	docker compose down

clean:
	@echo "Cleaning up TailSwan..."
	docker compose down -v
	docker rmi $(IMAGE_NAME):latest 2>/dev/null || true

logs:
	docker compose logs -f

shell:
	@echo "Opening shell in TailSwan container..."
	docker exec -it $(CONTAINER_NAME) /bin/bash

status:
	@echo "Checking TailSwan status..."
	docker exec -it $(CONTAINER_NAME) /tailswan/swan-status.sh status

rebuild: stop build run
	@echo "TailSwan rebuilt and restarted!"

tag:
	@echo "Tagging image for registry..."
	docker tag $(IMAGE_NAME):$(TAG) $(REGISTRY)/$(IMAGE_NAME):$(TAG)

push: build tag
	@echo "Pushing image to registry..."
	docker push $(REGISTRY)/$(IMAGE_NAME):$(TAG)
	@echo "Image pushed to $(REGISTRY)/$(IMAGE_NAME):$(TAG)"

go-build:
	@echo "Building Go binary..."
	go build -o bin/tailswan ./cmd/tailswan
	@echo "Binary built: bin/tailswan"

test:
	@echo "Running Go tests..."
	go test -v -race -coverprofile=coverage.out ./...

bin/golanci-lint: 
	@scripts/build-linter.sh

lint: bin/golanci-lint
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=5m

tidy:
	@echo "Running go mod tidy..."
	go mod tidy
	@git diff --exit-code go.mod go.sum || (echo "Error: go.mod or go.sum changed after 'go mod tidy'. Please commit the changes." && exit 1)
	@echo "go.mod and go.sum are tidy ✓"

fmt:
	@echo "Formatting Go code..."
	golangci-lint fmt
	@echo "Code formatted ✓"

install:
	@echo "Installing binary..."
	go install ./cmd/tailswan

ci: tidy lint test
	@echo "All CI checks passed ✓"
