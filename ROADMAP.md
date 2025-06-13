# IBKR Spread Automation Roadmap

## Project Overview

This roadmap breaks down the development of an automated vertical spread options trading system for Interactive Brokers into manageable phases. Following vibe coding principles, each phase maintains flow state while building toward a complete trading automation platform.

---

## Phase 0: Foundation & Environment Setup (Current Phase)

### Objectives
- Establish project structure following vibe coding principles
- Set up development environment with Docker
- Create core documentation framework

### Work Chunks

#### 0.1 Project Initialization
```gherkin
Feature: Project Foundation
  As a developer
  I want a well-structured project foundation
  So that development flows naturally

  Scenario: Initialize repository structure
    Given I have an empty project directory
    When I create the core directory structure
    Then I have src/, docs/, tests/, and docker/ directories
    And I have essential documentation files

  Scenario: Set up vibe coding artifacts
    Given I want to maintain creative flow
    When I create vibe-specific directories
    Then I have .vibe/, experiments/, and flow_journal/ directories
    And I have templates for capturing insights
```

#### 0.2 Docker Environment Setup
```gherkin
Feature: Containerized Development Environment
  As a developer
  I want Docker containers for each service
  So that development is reproducible and isolated

  Scenario: Create base Docker configurations
    Given I need three separate services
    When I create Docker configurations
    Then I have Dockerfiles for Python, Go scanner, and GUI
    And I have a docker-compose.yml orchestrating all services

  Scenario: Configure IBKR API requirements
    Given TWS requires specific network configuration
    When I set up Docker networking
    Then containers can access host TWS on port 7497/4001
    And I configure proper TCP socket connectivity
```

#### 0.3 Documentation Framework
- Create comprehensive ADR (Architecture Decision Record) structure
- Set up living documentation templates
- Initialize CHANGELOG.md and IDEAS.md

#### 0.4 TWS Configuration Setup
```gherkin
Feature: TWS Pre-Configuration
  As a developer
  I want clear TWS configuration requirements
  So that the API connection works properly

  Scenario: Configure TWS for API access
    Given I have TWS installed
    When I access Global Configuration
    Then I enable "ActiveX and Socket Clients"
    And I disable "Read-Only API"
    And I set memory allocation to 4000MB
    And I configure auto-restart time

  Scenario: Set up development environment
    Given I need both paper and live trading
    When I configure TWS
    Then I set up paper trading on port 7497
    And I document live trading port 7496
    And I configure trusted IP whitelist
```

### Deliverables
- [ ] Complete project directory structure
- [ ] Docker environment configurations
- [ ] Base documentation framework
- [ ] Development environment setup guide
- [ ] TWS configuration checklist
- [ ] API connection requirements documentation

---

## Phase 1: IBKR Connection Layer

### Objectives
- Establish reliable connection to Interactive Brokers TWS API
- Implement core trading operations wrapper
- Create robust error handling and retry logic

### Work Chunks

#### 1.1 Python IBKR Interface Container
```gherkin
Feature: IBKR API Connection
  As a trading system
  I want to connect to Interactive Brokers
  So that I can execute trades and receive market data

  Scenario: Establish TWS connection
    Given TWS is running on the host machine
    And TWS has "Enable ActiveX and Socket Clients" enabled
    And TWS has "Read-Only API" disabled
    When the Python container starts
    Then it connects to TWS API on port 7497 with unique clientId
    And it implements EReader thread for message handling
    And it maintains a stable connection with heartbeat

  Scenario: Handle connection failures
    Given a connection to TWS exists
    When the connection is interrupted (Error 1100)
    Then the system uses ib-insync Watchdog for auto-reconnection
    And it handles daily TWS restart at configured time
    And it logs all connection events with error codes

  Scenario: Manage API rate limits
    Given TWS has pacing limitations (50 req/sec default)
    When the system makes API requests
    Then it respects the pacing limit to avoid Error 100
    And it implements request throttling (45 req/sec safety)
    And it handles pacing violations gracefully
```

#### 1.2 Core Trading Operations
```gherkin
Feature: Trading Operations Wrapper
  As a trading system
  I want abstracted trading operations
  So that higher-level services can execute trades simply

  Scenario: Place vertical spread order
    Given I have validated spread parameters
    When I request spread execution
    Then the system creates a combo order
    And it submits to IBKR with proper legs

  Scenario: Monitor order status
    Given I have submitted an order
    When the order status changes
    Then the system captures the update
    And it notifies interested services
```

#### 1.3 Market Data Streaming
- Implement real-time options chain data retrieval via reqSecDefOptParams()
- Handle market data line limits based on subscription level
- Create efficient data caching layer to reduce API calls
- Set up Greeks calculation pipeline with reqMktData()
- Implement tick-by-tick data handling for precise fills
- Manage multiple data subscriptions within TWS limits

### Deliverables
- [ ] Working Python container with ib-insync v0.9.86
- [ ] Connection management system with Watchdog auto-recovery
- [ ] Basic trading operations API with error handling
- [ ] Market data streaming capability within rate limits
- [ ] TWS configuration validation and setup guide
- [ ] Request pacing and throttling implementation

---

## Phase 2: Options Scanner Engine

### Objectives
- Build high-performance options scanning in Go
- Implement all parameter filters from vision
- Create efficient data processing pipeline

### Work Chunks

#### 2.1 Scanner Core Architecture
```gherkin
Feature: Options Scanner Foundation
  As a trader
  I want fast, accurate options scanning
  So that I can identify opportunities quickly

  Scenario: Scan with basic filters
    Given I have market data for SPY options
    When I apply delta and DTE filters
    Then the scanner returns matching contracts
    And results are returned within 100ms

  Scenario: Handle high-volume scanning
    Given I'm scanning 50+ underlying symbols
    When multiple scans run concurrently
    Then the system maintains performance
    And memory usage remains stable
```

#### 2.2 Advanced Filter Implementation
```gherkin
Feature: Comprehensive Filter Suite
  As an experienced trader
  I want all standard options parameters available
  So that I can implement sophisticated strategies

  Scenario: Apply Greeks-based filters
    Given I set delta range 0.25-0.35
    And I set theta greater than -0.05
    When scanning for debit spreads
    Then only contracts matching all Greeks criteria appear

  Scenario: IV percentile filtering
    Given current IV percentile is 85%
    When I set minimum IV percentile to 80%
    Then the scanner switches to credit spread mode
    And it identifies high IV opportunities
```

#### 2.3 Scanner Optimization
- Implement concurrent scanning with goroutines
- Add caching for frequently accessed data
- Create benchmarking suite

### Deliverables
- [ ] Go scanner container with API
- [ ] Complete filter implementation
- [ ] Performance benchmarks
- [ ] Scanner configuration system

---

## Phase 3: GUI Development

### Objectives
- Create intuitive Windows application
- Implement real-time visualization
- Build comprehensive parameter control panel

### Work Chunks

#### 3.1 GUI Foundation
```gherkin
Feature: Desktop Application Framework
  As a trader
  I want a responsive desktop interface
  So that I can control my trading system efficiently

  Scenario: Launch application
    Given the Docker containers are running
    When I start the Windows application
    Then it connects to backend services
    And displays connection status

  Scenario: Navigate between views
    Given I'm in the scanner view
    When I switch to positions view
    Then the transition is smooth
    And my scanner continues running
```

#### 3.2 Parameter Control Implementation
```gherkin
Feature: Granular Parameter Control
  As a power user
  I want fine-grained control over all parameters
  So that I can tune my strategy precisely

  Scenario: Adjust filter parameters
    Given I'm viewing the parameter panel
    When I adjust the delta slider
    Then the scanner updates in real-time
    And I see results refresh immediately

  Scenario: Save parameter profiles
    Given I've configured my parameters
    When I save as "High IV Strategy"
    Then the profile persists
    And I can load it in future sessions
```

#### 3.3 Real-time Visualization
- Implement live scanner results grid
- Create options chain visualization
- Add P&L tracking dashboard

### Deliverables
- [ ] Working Svelte frontend
- [ ] Go backend API layer
- [ ] Real-time WebSocket updates
- [ ] Parameter persistence system

---

## Phase 4: Integration & Orchestration

### Objectives
- Unite all components into cohesive system
- Implement inter-service communication
- Create deployment configuration

### Work Chunks

#### 4.1 Service Integration
```gherkin
Feature: Multi-Service Coordination
  As a complete system
  I want services to work together seamlessly
  So that traders experience unified functionality

  Scenario: End-to-end trade execution
    Given all services are running
    And TWS is authenticated and connected
    When I click "Execute Trade" in GUI
    Then scanner identifies opportunities
    And Python service validates order via whatIfOrder()
    And places combo order with proper OCA group
    And monitors order status via callbacks
    And GUI shows real-time order updates

  Scenario: Handle TWS daily restart
    Given TWS restarts daily at configured time
    When the restart window approaches
    Then the system gracefully disconnects
    And waits for TWS to become available
    And automatically reconnects with saved state

  Scenario: Handle service failures
    Given a service becomes unavailable
    When other services detect the failure
    Then they enter degraded mode gracefully
    And alert the user appropriately
```

#### 4.2 Docker Compose Orchestration
- Create production-ready docker-compose.yml
- Configure host networking for TWS connection (127.0.0.1:7497)
- Implement health checks for all services
- Set up inter-container networking
- Handle TWS/IB Gateway memory requirements (4GB recommended)
- Configure container restart policies for TWS daily restart

### Deliverables
- [ ] Complete docker-compose configuration
- [ ] Service discovery mechanism
- [ ] Integration test suite
- [ ] Deployment documentation

---

## Phase 5: Testing & Hardening

### Objectives
- Ensure system reliability
- Implement comprehensive error handling
- Create monitoring and logging

### Work Chunks

#### 5.1 Testing Framework
```gherkin
Feature: Comprehensive Testing
  As a trading system
  I want thorough testing coverage
  So that traders can rely on the system

  Scenario: Simulate market conditions
    Given historical market data
    When I run backtesting scenarios
    Then the system behaves predictably
    And no orders exceed risk parameters

  Scenario: Test API rate limit compliance
    Given TWS pacing limit of 50 requests/second
    When the system runs at full capacity
    Then it never exceeds 45 requests/second
    And handles Error 100 (pacing violation) gracefully

  Scenario: Stress test scanner
    Given market data line limits per TWS subscription
    When multiple scans run concurrently
    Then the system respects subscription limits
    And queues excess requests appropriately
```

#### 5.2 Error Handling & Recovery
- Implement circuit breakers for API errors
- Handle all TWS error codes (502, 507, 100, 1100, etc.)
- Add comprehensive logging with TWS message log integration
- Create system monitoring dashboard
- Implement reconnection logic for Error 1100 and socket errors
- Handle contract validation errors gracefully

### Deliverables
- [ ] Integration test suite
- [ ] Performance test results
- [ ] Error handling documentation
- [ ] Monitoring setup

---

## Phase 6: Advanced Features

### Objectives
- Implement sophisticated trading features
- Add machine learning capabilities
- Create advanced analytics

### Work Chunks

#### 6.1 Advanced Trading Logic
```gherkin
Feature: Sophisticated Trading Strategies
  As an advanced trader
  I want complex strategy capabilities
  So that I can maximize opportunities

  Scenario: Implement rolling logic
    Given I have an open position near expiration
    When rollover conditions are met
    Then the system suggests optimal roll
    And can execute automatically if enabled

  Scenario: Multi-leg strategies
    Given I want iron condor capability
    When I enable 4-leg strategies
    Then the scanner identifies opportunities
    And can execute complex orders
```

#### 6.2 Analytics & Reporting
- Build performance analytics dashboard
- Create trade journal export
- Implement strategy backtesting

### Deliverables
- [ ] Advanced strategy modules
- [ ] Analytics dashboard
- [ ] Backtesting framework
- [ ] Export capabilities

---

## Phase 7: Production Readiness

### Objectives
- Prepare for live trading
- Implement security measures
- Create deployment pipeline

### Work Chunks

#### 7.1 Security Implementation
- Add authentication system
- Configure TWS Trusted IP whitelist
- Implement secure clientId management
- Create audit logging for all trades
- Handle TWS two-factor authentication flow
- Secure storage of TWS connection parameters

#### 7.2 Deployment Pipeline
- Set up CI/CD workflows
- Create installation package
- Write operations runbook

### Deliverables
- [ ] Security documentation
- [ ] Installer package
- [ ] Operations guide
- [ ] Production checklist

---

## Success Metrics

### Technical Requirements Summary

Based on IBKR TWS API documentation:
- **Connection**: TCP socket to 127.0.0.1:7497 (TWS) or 4001 (Gateway)
- **Threading**: Minimum 2 threads (send + EReader)
- **Rate Limits**: 50 req/sec default (45 req/sec safe threshold)
- **Authentication**: Manual TWS login required (no headless)
- **Daily Restart**: TWS restarts daily at configured time
- **Memory**: 4GB recommended for TWS with API
- **Error Handling**: Comprehensive error code handling required
- **Data Limits**: Based on market data subscription level

### Phase Completion Criteria
- All Gherkin scenarios pass
- Documentation is current
- Code follows vibe coding principles
- Integration tests pass
- Performance benchmarks met

### Project Success Indicators
- Sub-second scanner response times
- 99.9% order execution reliability
- Intuitive GUI with <3 clicks to any feature
- Complete parameter coverage from vision
- Positive trader feedback on ease of use

---

## Vibe Coding Principles Applied

Throughout all phases:
- Maintain flow state with clear, focused work chunks
- Document decisions as they're made (ADRs)
- Keep experimental branches for creative exploration
- Regular flow journal entries capturing insights
- Flexible timeline allowing for creative discovery
- Living documentation that evolves with the project

---

## Next Steps

1. Complete Phase 0 foundation work
2. Set up development environment
3. Begin Phase 1 with IBKR connection prototype
4. Maintain IDEAS.md for future enhancements
5. Regular retrospectives to adjust roadmap as needed