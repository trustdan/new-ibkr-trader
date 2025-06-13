# Changelog

All notable changes to the IBKR Spread Automation project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project structure with vibe coding principles
- Docker-based microservices architecture
- Python IBKR interface using ib-insync
- Go high-performance scanner engine
- Svelte-based GUI application
- Comprehensive monitoring with Prometheus/Grafana
- Event-driven async architecture throughout

### Architecture Decisions
- Chose async/event-driven pattern for Python service (ADR-001)
- Implemented subscription management with LRU eviction
- Added request coordination between services

### Developer Experience
- Flow journal system for maintaining creative momentum
- Experiments folder for safe exploration
- Living documentation approach
- Vibe-friendly project organization

---

## Version Format
`[version] - YYYY-MM-DD`

### Types of changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` in case of vulnerabilities

### Vibe Notes
Each version should include a brief "vibe check" - how did this release feel to develop?