# Makefile with async-aware commands
.PHONY: dev test monitor clean vibe paper-test

dev:
	@echo "🚀 Starting async development environment..."
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

help:
	@echo "Available commands:"
	@echo "  make dev        - Start development environment"
	@echo "  make test       - Run async test suite"
	@echo "  make paper-test - Run paper trading validation"
	@echo "  make monitor    - Open monitoring dashboards"
	@echo "  make vibe       - Check the vibe"
	@echo "  make logs       - Tail service logs"
	@echo "  make clean      - Clean up containers"
	@echo "  make rebuild    - Rebuild containers"