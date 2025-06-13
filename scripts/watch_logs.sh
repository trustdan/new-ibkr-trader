#!/bin/bash
# Watch logs from multiple services in split terminal

# Colors for different services
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if docker-compose is running
if ! docker-compose ps | grep -q "Up"; then
    echo "⚠️  No services running. Start with: make dev"
    exit 1
fi

# Function to colorize logs
colorize_logs() {
    service=$1
    color=$2
    docker-compose logs -f --tail=50 $service 2>&1 | while IFS= read -r line; do
        echo -e "${color}[$service]${NC} $line"
    done
}

echo "📜 Watching logs from all services..."
echo "Press Ctrl+C to stop"
echo "================================"

# Watch logs from different services with different colors
colorize_logs "python-ibkr" "$GREEN" &
colorize_logs "go-scanner" "$BLUE" &
colorize_logs "gui-backend" "$YELLOW" &

# Wait for all background processes
wait