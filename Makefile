# Makefile with async-aware commands and Docker support
.PHONY: dev test monitor clean vibe paper-test docker-build docker-push

# Development commands
dev:
	@echo "🚀 Starting async development environment..."
	docker-compose -f docker-compose.dev.yml up
	
dev-detached:
	@echo "🚀 Starting development environment in background..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "⏳ Waiting for services..."
	@sleep 5
	@make health-check

# Production commands
prod:
	@echo "🏭 Starting production environment..."
	docker-compose up -d
	@echo "⏳ Waiting for services..."
	@sleep 5
	@make health-check

test:
	@echo "🧪 Running async tests..."
	docker-compose run --rm python-ibkr pytest -v --asyncio-mode=auto

paper-test:
	@echo "📄 Running paper trading validation suite..."
	docker-compose -f docker-compose.paper.yml up -d
	@sleep 10
	@docker-compose run --rm test-runner pytest tests/paper_trading/ -v

monitor:
	@echo "📊 Opening monitoring dashboards..."
	open http://localhost:9090  # Prometheus
	open http://localhost:3000  # Grafana

vibe:
	@echo "🌊 Checking the vibe..."
	@cat .vibe/manifesto.md
	@echo "\n📝 Latest flow journal entry:"
	@ls -t flow_journal/*.md | head -1 | xargs tail -20

health-check:
	@docker-compose ps
	@curl -s http://localhost:8080/health | jq

clean:
	@echo "🧹 Cleaning up containers and volumes..."
	docker-compose down -v
	@echo "✨ Clean complete!"

logs:
	@echo "📜 Tailing service logs..."
	docker-compose logs -f --tail=100

rebuild:
	@echo "🔨 Rebuilding containers..."
	docker-compose build --no-cache
	@echo "✅ Rebuild complete!"

# Docker commands
docker-build:
	@echo "🐳 Building Docker images locally..."
	docker-compose -f docker-compose.dev.yml build

docker-push:
	@echo "🐳 Pushing images to Docker Hub..."
	@echo "⚠️  Make sure to set DOCKER_REGISTRY environment variable"
	docker-compose build
	docker-compose push

docker-pull:
	@echo "🐳 Pulling latest images from Docker Hub..."
	docker-compose pull

# Utility commands
shell-python:
	@echo "🐍 Opening Python shell..."
	docker-compose -f docker-compose.dev.yml run --rm python /bin/bash

shell-scanner:
	@echo "🔍 Opening Scanner shell..."
	docker-compose -f docker-compose.dev.yml run --rm scanner /bin/sh

help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Start development environment (foreground)"
	@echo "  make dev-detached - Start development environment (background)"
	@echo "  make test         - Run async test suite"
	@echo "  make paper-test   - Run paper trading validation"
	@echo ""
	@echo "Production:"
	@echo "  make prod         - Start production environment"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker images locally"
	@echo "  make docker-push  - Push images to Docker Hub"
	@echo "  make docker-pull  - Pull latest images"
	@echo ""
	@echo "Utilities:"
	@echo "  make monitor      - Open monitoring dashboards"
	@echo "  make vibe         - Check the vibe"
	@echo "  make logs         - Tail service logs"
	@echo "  make clean        - Clean up containers"
	@echo "  make rebuild      - Rebuild containers"
	@echo "  make shell-python - Open Python container shell"
	@echo "  make shell-scanner- Open Scanner container shell"