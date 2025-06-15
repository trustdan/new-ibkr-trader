# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**IMPORTANT**: Always start each session by reading RULES.md for quick reference and vibe coding principles.

## Session Initialization

- At the start of each session, please review the RULES.md, MASTER_PLAN_UNIFIED.md, flow_journal/, flow-journal/, logs, and commit history in order to see where we're at in the grand scheme of things

## Project Overview

**IBKR Spread Automation** - A Windows application for automating vertical spread options trading through Interactive Brokers.

## Architecture

This project uses a microservices architecture with Docker containers:

1. **Python Interface Container**: Handles IBKR TWS API communication using the `ib-insync` library (v0.9.86)
2. **Go Scanner Container**: High-performance options scanning based on customizable criteria  
3. **GUI Windows App**: Go backend with Svelte frontend for user interaction

## Key Libraries and APIs

- **ib-insync**: Python wrapper for Interactive Brokers TWS API
- **Interactive Brokers TWS API**: For live trading and market data
- **Docker Compose**: For container orchestration

## Development Commands

Since this project is in the documentation phase, specific commands will be added as the implementation progresses. Expected commands will include:

```bash
# Docker operations (to be implemented)
docker-compose up      # Start all services
docker-compose down    # Stop all services

# Python container (to be implemented)
pip install ib-insync  # Install IBKR Python wrapper
python src/ib_connector.py  # Run IBKR connection service

# Go scanner (to be implemented)
go run cmd/scanner/main.go  # Run options scanner

# GUI development (to be implemented)
npm run dev     # Start Svelte development server
npm run build   # Build production GUI
```

## Project Structure

```
/
├── claude-instructions/    # AI development guidelines
│   ├── VISION.md          # Project vision and requirements
│   └── vibe-instruct.md   # Development methodology
├── documentation/         # API and library documentation
│   ├── IBKR-TWS-API-documentation.md
│   └── ib-insync*.txt    # ib-insync library docs
├── src/                  # (To be created) Source code
│   ├── python/          # Python IBKR interface
│   ├── go/              # Go scanner and backend
│   └── gui/             # Svelte frontend
└── docker-compose.yml    # (To be created) Container orchestration
```

## Key Trading Parameters

The application handles extensive options trading parameters including:

- Liquidity metrics (open interest, volume)
- Greeks (delta, gamma, theta, vega)
- Implied volatility levels and percentiles
- Days to expiration (DTE) ranges
- Bid-ask spreads
- Probability metrics (ITM, PoP)
- Position and risk management limits

## Development Philosophy

This project follows "vibe coding" principles emphasizing:
- Flow state preservation
- Intuitive code organization
- Living documentation that evolves with the project
- Comprehensive documentation before implementation

## Important Notes

1. The project is currently in the documentation/planning phase
2. All trading logic should prioritize risk management and error handling
3. The GUI should provide real-time feedback for all scanner operations
4. Docker containers should be independently scalable
5. Always validate trading parameters before executing orders through IBKR

## Testing Strategy

(To be implemented) The project will require:
- Unit tests for trading logic
- Integration tests for IBKR API communication
- End-to-end tests for the complete trading workflow
- Mock trading environment for safe testing