# Vision.md

## Project Name

**IBKR Spread Automation**

---

## Vision

Create an intuitive, robust, and efficient Windows application to automate vertical spread options trading (primarily debit spreads, switching to credit spreads when implied volatility is high) through Interactive Brokers. This application empowers experienced traders by providing granular, highly customizable contract selection and execution control across all typical options parameters.

---

## Components

### Dockerized Containers

1. **Python Interface Container**: Connects and communicates directly with Interactive Brokers' Trader Workstation (TWS) API, executing orders, managing positions, and monitoring real-time trading data.
2. **Go Scanner Container**: Scans and identifies optimal trading opportunities based on extensive, customizable criteria covering all key options metrics.
3. **GUI Windows App (Go & Svelte)**: A responsive interface for adjusting and fine‑tuning trade parameters, visualizing scanner results, and managing strategy workflows.

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

---

## Target User

Experienced options traders who demand precision, speed, and deep customization in automated spread trading.

---

## Success Criteria

* Automated trades consistently execute within trader‑defined parameter bounds.
* GUI responsiveness under heavy scanning workloads.
* Trader satisfaction with ease of parameter tuning and strategy management.

---

Unleash the full power of automated vertical spreads by giving traders **complete control** over every aspect of their methodology.

