.PHONY: build run stop clean logs shell help

IMAGE_NAME := tailswan
CONTAINER_NAME := tailswan

help:
	@echo "TailSwan Makefile Commands"
	@echo ""
	@echo "  make build       - Build the Docker image"
	@echo "  make run         - Run the container with docker-compose"
	@echo "  make stop        - Stop the container"
	@echo "  make clean       - Stop and remove container and volumes"
	@echo "  make logs        - Show container logs"
	@echo "  make shell       - Open a shell in the running container"
	@echo "  make status      - Show TailSwan status"
	@echo "  make rebuild     - Rebuild and restart the container"
	@echo ""

build:
	@echo "Building TailSwan Docker image..."
	docker build -t $(IMAGE_NAME):latest .

run:
	@echo "Starting TailSwan..."
	docker-compose up -d

stop:
	@echo "Stopping TailSwan..."
	docker-compose down

clean:
	@echo "Cleaning up TailSwan..."
	docker-compose down -v
	docker rmi $(IMAGE_NAME):latest 2>/dev/null || true

logs:
	docker-compose logs -f

shell:
	@echo "Opening shell in TailSwan container..."
	docker exec -it $(CONTAINER_NAME) /bin/bash

status:
	@echo "Checking TailSwan status..."
	docker exec -it $(CONTAINER_NAME) /tailswan/swan-status.sh status

rebuild: stop build run
	@echo "TailSwan rebuilt and restarted!"
