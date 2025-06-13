# Vision.md

## Project Name

**IBKR Spread Automation**

---

## Vision

Create an intuitive, robust, and efficient Windows application to automate vertical spread options trading (primarily debit spreads, switching to credit spreads when implied volatility is high) through Interactive Brokers. This application empowers experienced traders by providing granular, highly customizable contract selection and execution control across all typical options parameters.

---

## Components

### Dockerized Containers

1. **Python Interface Container**: 
   - Connects to TWS API via TCP socket (127.0.0.1:7497 or IB Gateway on 4001)
   - Uses ib-insync v0.9.86 for simplified async API handling
   - Implements EReader thread for message processing
   - Manages connection lifecycle including daily TWS restarts
   - Handles API rate limiting (45 req/sec safe threshold)
   - Executes combo orders for vertical spreads with OCA groups
   - Monitors real-time position updates and order status callbacks

2. **Go Scanner Container**: 
   - High-performance concurrent scanning with goroutines
   - Respects market data line limits based on TWS subscription
   - Implements request queuing to avoid pacing violations
   - Caches frequently accessed contract details
   - Identifies optimal trading opportunities based on extensive criteria

3. **GUI Windows App (Go & Svelte)**: 
   - Responsive interface requiring no more than 3 clicks to any feature
   - Real-time WebSocket updates for scanner results and positions
   - Comprehensive parameter control panel with persistence
   - Visual indicators for connection status and API health
   - Trade execution confirmation with whatIfOrder() preview

---

## Key Selection Parameters

Options traders often fine‑tune:

* **Liquidity & Volume**: minimum open interest; minimum average daily option volume
* **Bid‑Ask Spread**: maximum allowable spread to ensure tight pricing
* **Underlying Price Filters**: absolute price range or percentage moves
* **Moneyness**: ITM / ATM / OTM thresholds
* **Days to Expiration (DTE)**: range of DTE or specific expiration cycles (weekly, monthly)
* **Greeks Thresholds**: delta, gamma, theta, vega, rho ranges or target sensitivities
* **Implied Volatility (IV)**: IV level, IV percentile or rank versus historical
* **Probability Metrics**: probability ITM, probability of profit (PoP)
* **Spread Width**: minimum and maximum strike distance
* **Risk & Reward Limits**: max debit/credit; max risk per contract; target ROI
* **Event & Earnings Filters**: exclude or include contracts near earnings, dividends, or other corporate events
* **Technical / Fundamental Filters** (optional): SMA, RSI, volatility breakouts, sector filters
* **Position & Capital Management**: max open positions; maximum capital usage; margin impact
* **Assignment & Early Exercise Risk**: thresholds for risk of assignment
* **Order Types & Time‑In‑Force**: limit vs market; GTC, DAY, IOC, FOK
* **Rolling & Adjustment Rules**: criteria for rolling out or adjusting existing spreads

---

## User Stories (Gherkin)

```gherkin
Feature: Fine‑Tune Options Filters
  As an options trader
  I want to specify variable ranges for all standard options parameters
  So that automated trades reflect my detailed strategy

  Scenario: Filter by delta and DTE
    Given I set delta between 0.20 and 0.35
    And I set DTE between 30 and 60
    When the scanner runs
    Then I see only contracts matching these criteria

  Scenario: Limit bid‑ask spread
    Given I set max bid-ask spread to 0.10
    When scanning options
    Then only contracts with spread ≤ 0.10 are returned

  Scenario: Target high IV percentile
    Given I set IV percentile ≥ 80%
    When IV is elevated
    Then the scanner chooses credit spreads
```

---

## Essential Features

1. **GUI Parameter Panel** with sliders, dropdowns, and inputs for every key filter above.
2. **Real‑Time Scanner Visualization** showing top candidates, Greeks, IV, and probability metrics.
3. **Trade Execution Log** with confirmations, errors, and audit trail.
4. **Strategy Profiles** to save and load multiple parameter sets and workflows.
5. **Container Orchestration** using Docker Compose for seamless local development and deployment.
6. **Robust Error Handling** and retry logic in both Python and Go services.

## Technical Requirements (IBKR TWS API)

1. **Connection Management**:
   - TWS must have "Enable ActiveX and Socket Clients" enabled
   - "Read-Only API" must be disabled for trading
   - Unique clientId per connection (max 32 concurrent)
   - Automatic handling of daily TWS restart
   - Watchdog implementation for connection recovery

2. **Rate Limiting & Performance**:
   - Respect 50 requests/second pacing limit (45 req/sec recommended)
   - Handle Error 100 (pacing violation) with exponential backoff
   - Queue requests when approaching limits
   - Implement request batching where possible

3. **Order Execution**:
   - Use combo orders for vertical spreads
   - Implement OCA (One-Cancels-All) groups for risk management
   - Preview orders with whatIfOrder() for margin impact
   - Monitor order status via asynchronous callbacks

4. **Data Management**:
   - Subscribe to real-time market data within line limits
   - Cache contract details to reduce API calls
   - Handle option chain requests efficiently
   - Implement proper cleanup on disconnection

5. **Error Handling**:
   - Comprehensive handling of all TWS error codes
   - Automatic reconnection on Error 1100 (connectivity lost)
   - Graceful degradation when services unavailable
   - Detailed logging of all API interactions

---

## Target User

Experienced options traders who demand precision, speed, and deep customization in automated spread trading.

---

## Success Criteria

* Automated trades consistently execute within trader‑defined parameter bounds.
* GUI responsiveness under heavy scanning workloads.
* Trader satisfaction with ease of parameter tuning and strategy management.
* System maintains stable TWS connection through daily restarts.
* Zero pacing violations during normal operation.
* Sub-second order execution from signal to TWS confirmation.
* 99.9% uptime excluding scheduled TWS maintenance windows.

---

Unleash the full power of automated vertical spreads by giving traders **complete control** over every aspect of their methodology.

