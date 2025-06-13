# Scanner Filter Implementations Guide

## Overview

The scanner employs a comprehensive suite of filters to identify optimal vertical spread opportunities. Each filter is designed for maximum performance while providing the flexibility traders need to implement their strategies.

## Filter Categories

### 1. Time-Based Filters
- **DTE (Days to Expiration) Filter**
- **Expiration Date Filter**

### 2. Greeks-Based Filters
- **Delta Filter**
- **Gamma Filter**
- **Theta Filter**
- **Vega Filter**

### 3. Liquidity Filters
- **Volume Filter**
- **Open Interest Filter**
- **Bid-Ask Spread Filter**

### 4. Price & Probability Filters
- **Strike Price Filter**
- **Spread Width Filter**
- **Probability of Profit (PoP) Filter**
- **In-The-Money (ITM) Probability Filter**

### 5. Volatility Filters
- **Implied Volatility (IV) Filter**
- **IV Percentile Filter**
- **IV Rank Filter**

## Filter Implementations

### DTE (Days to Expiration) Filter

```go
type DTEFilter struct {
    MinDays int
    MaxDays int
}

func (f *DTEFilter) Apply(contracts []Contract) []Contract {
    now := time.Now()
    filtered := make([]Contract, 0, len(contracts)/2)
    
    for _, contract := range contracts {
        dte := int(contract.Expiry.Sub(now).Hours() / 24)
        if dte >= f.MinDays && dte <= f.MaxDays {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *DTEFilter) Selectivity() float64 {
    // Typically filters 70-80% of contracts
    return 0.75
}
```

**Use Cases:**
- Short-term trades: 7-45 DTE
- Monthly income: 30-60 DTE
- LEAPS strategies: 180+ DTE

### Delta Filter

```go
type DeltaFilter struct {
    MinDelta float64
    MaxDelta float64
    UseAbsolute bool  // For puts, use absolute value
}

func (f *DeltaFilter) Apply(contracts []Contract) []Contract {
    filtered := make([]Contract, 0, len(contracts)/3)
    
    for _, contract := range contracts {
        delta := contract.Greeks.Delta
        if f.UseAbsolute && contract.Right == "P" {
            delta = math.Abs(delta)
        }
        
        if delta >= f.MinDelta && delta <= f.MaxDelta {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *DeltaFilter) Priority() int {
    // Apply after DTE but before complex Greeks
    return 20
}
```

**Common Delta Ranges:**
- Conservative: 0.25-0.35
- Moderate: 0.30-0.40
- Aggressive: 0.40-0.50
- Deep ITM: > 0.70

### Liquidity Filter

```go
type LiquidityFilter struct {
    MinVolume       int
    MinOpenInterest int
    MaxBidAskSpread float64
    RequireBothLegs bool  // For spreads
}

func (f *LiquidityFilter) Apply(contracts []Contract) []Contract {
    filtered := make([]Contract, 0, len(contracts)/4)
    
    for _, contract := range contracts {
        hasVolume := contract.Volume >= f.MinVolume
        hasOI := contract.OpenInterest >= f.MinOpenInterest
        tightSpread := (contract.Ask - contract.Bid) <= f.MaxBidAskSpread
        
        if (hasVolume || hasOI) && tightSpread {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *LiquidityFilter) Selectivity() float64 {
    // Most selective filter - typically keeps only 20-30%
    return 0.25
}
```

**Liquidity Thresholds:**
- High Liquidity: Volume > 100, OI > 500
- Medium Liquidity: Volume > 50, OI > 100
- Low Liquidity: Volume > 10, OI > 50

### Greeks Composite Filter

```go
type GreeksFilter struct {
    MaxGamma    float64
    MinTheta    float64  // Negative for long options
    MaxVega     float64
    ThetaGamma  float64  // Theta/Gamma ratio
}

func (f *GreeksFilter) Apply(contracts []Contract) []Contract {
    filtered := make([]Contract, 0, len(contracts)/2)
    
    for _, contract := range contracts {
        g := contract.Greeks
        
        // Basic Greeks checks
        if g.Gamma > f.MaxGamma {
            continue
        }
        if g.Theta < f.MinTheta {
            continue
        }
        if g.Vega > f.MaxVega {
            continue
        }
        
        // Advanced ratio check
        if f.ThetaGamma > 0 && g.Gamma != 0 {
            ratio := math.Abs(g.Theta / g.Gamma)
            if ratio < f.ThetaGamma {
                continue
            }
        }
        
        filtered = append(filtered, contract)
    }
    
    return filtered
}
```

**Greeks Guidelines:**
- Gamma Risk: Max 0.05 for conservative
- Theta Decay: Min -0.10 for income strategies
- Vega Exposure: Max 0.50 for low volatility risk

### IV Percentile Filter

```go
type IVPercentileFilter struct {
    MinPercentile int
    MaxPercentile int
    LookbackDays  int
    cache         *IVHistoryCache
}

func (f *IVPercentileFilter) Apply(contracts []Contract) []Contract {
    filtered := make([]Contract, 0, len(contracts)/2)
    
    for _, contract := range contracts {
        percentile := f.calculateIVPercentile(contract)
        
        if percentile >= f.MinPercentile && percentile <= f.MaxPercentile {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *IVPercentileFilter) calculateIVPercentile(contract Contract) int {
    // Get historical IV data from cache
    history := f.cache.GetHistory(contract.Symbol, f.LookbackDays)
    if len(history) == 0 {
        return 50 // Default to median if no history
    }
    
    // Calculate percentile
    currentIV := contract.ImpliedVol
    rank := 0
    for _, historicalIV := range history {
        if currentIV > historicalIV {
            rank++
        }
    }
    
    return (rank * 100) / len(history)
}
```

**IV Percentile Strategies:**
- High IV (> 70th percentile): Credit spreads
- Low IV (< 30th percentile): Debit spreads
- Median IV (40-60th percentile): Iron condors

### Spread Width Filter

```go
type SpreadWidthFilter struct {
    MinWidth float64
    MaxWidth float64
    AsPercent bool  // Width as % of stock price
}

func (f *SpreadWidthFilter) ApplyToSpreads(spreads []VerticalSpread) []VerticalSpread {
    filtered := make([]VerticalSpread, 0, len(spreads)/2)
    
    for _, spread := range spreads {
        width := math.Abs(spread.LongStrike - spread.ShortStrike)
        
        if f.AsPercent && spread.UnderlyingPrice > 0 {
            width = (width / spread.UnderlyingPrice) * 100
        }
        
        if width >= f.MinWidth && width <= f.MaxWidth {
            filtered = append(filtered, spread)
        }
    }
    
    return filtered
}
```

**Width Guidelines:**
- Narrow Spreads: $5-10 or 2-5%
- Standard Spreads: $10-25 or 5-10%
- Wide Spreads: $25+ or 10%+

### Probability Filters

```go
type ProbabilityFilter struct {
    MinPoP float64  // Probability of Profit
    MaxITM float64  // In-The-Money probability
}

func (f *ProbabilityFilter) Apply(contracts []Contract) []Contract {
    filtered := make([]Contract, 0, len(contracts)/2)
    
    for _, contract := range contracts {
        pop := f.calculatePoP(contract)
        itm := f.calculateITM(contract)
        
        if pop >= f.MinPoP && itm <= f.MaxITM {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}

func (f *ProbabilityFilter) calculatePoP(contract Contract) float64 {
    // Using delta as proxy for probability
    if contract.Right == "C" {
        return 1.0 - math.Abs(contract.Greeks.Delta)
    } else {
        return math.Abs(contract.Greeks.Delta)
    }
}

func (f *ProbabilityFilter) calculateITM(contract Contract) float64 {
    // More sophisticated calculation using Black-Scholes
    return contract.ProbITM // Pre-calculated by Python service
}
```

**Probability Targets:**
- High Probability: PoP > 70%
- Balanced: PoP 50-70%
- High Risk/Reward: PoP 30-50%

## Caching Strategies

### Filter Result Caching

```go
type FilterCache struct {
    cache    map[string]*CacheEntry
    ttl      time.Duration
    maxSize  int
    mu       sync.RWMutex
}

type CacheEntry struct {
    Results   []Contract
    Timestamp time.Time
    HitCount  int
}

func (fc *FilterCache) GetOrCompute(
    key string, 
    compute func() []Contract,
    ttl time.Duration,
) []Contract {
    fc.mu.RLock()
    if entry, ok := fc.cache[key]; ok {
        if time.Since(entry.Timestamp) < ttl {
            entry.HitCount++
            fc.mu.RUnlock()
            return entry.Results
        }
    }
    fc.mu.RUnlock()
    
    // Compute and cache
    results := compute()
    fc.Set(key, results, ttl)
    return results
}
```

### Cache Key Generation

```go
func generateCacheKey(filter Filter, contracts []Contract) string {
    h := fnv.New64a()
    
    // Include filter parameters
    filterBytes, _ := json.Marshal(filter)
    h.Write(filterBytes)
    
    // Include contract identifiers (not full data)
    for _, c := range contracts[:min(10, len(contracts))] {
        h.Write([]byte(c.Symbol))
        h.Write([]byte(c.Expiry.Format("20060102")))
    }
    
    return fmt.Sprintf("%s_%x", filter.Name(), h.Sum64())
}
```

### TTL Strategy

Different filters require different cache TTLs:

| Filter Type | TTL | Reason |
|------------|-----|---------|
| DTE | 24 hours | Changes daily |
| Greeks | 5 minutes | Frequent updates |
| Liquidity | 1 minute | Real-time critical |
| IV Percentile | 1 hour | Historical data |
| Static (Strike) | Until expiry | Never changes |

## Filter Chain Optimization

### Dynamic Reordering

```go
type FilterOptimizer struct {
    stats map[string]*FilterStats
}

type FilterStats struct {
    AvgSelectivity   float64
    AvgExecutionTime time.Duration
    TotalRuns        int
}

func (fo *FilterOptimizer) OptimizeOrder(filters []Filter) []Filter {
    // Sort by selectivity * execution_time
    // Most selective and fastest filters first
    sort.Slice(filters, func(i, j int) bool {
        statsI := fo.stats[filters[i].Name()]
        statsJ := fo.stats[filters[j].Name()]
        
        scoreI := statsI.AvgSelectivity * float64(statsI.AvgExecutionTime)
        scoreJ := statsJ.AvgSelectivity * float64(statsJ.AvgExecutionTime)
        
        return scoreI < scoreJ
    })
    
    return filters
}
```

### Parallel Filter Groups

```go
func groupIndependentFilters(filters []Filter) [][]Filter {
    groups := [][]Filter{
        // Group 1: Time-based (independent)
        {dteFilter, expiryFilter},
        
        // Group 2: Price-based (independent)
        {strikeFilter, spreadWidthFilter},
        
        // Group 3: Greeks (may depend on each other)
        {deltaFilter, gammaFilter, thetaFilter, vegaFilter},
        
        // Group 4: Market data dependent
        {liquidityFilter, ivFilter},
    }
    
    return groups
}
```

## Performance Considerations

### Memory Management

```go
// Pre-allocate slices with expected capacity
filtered := make([]Contract, 0, len(contracts)/expectedSelectivity)

// Reuse slices when possible
var contractPool = sync.Pool{
    New: func() interface{} {
        return make([]Contract, 0, 1000)
    },
}
```

### Benchmarking

```go
func BenchmarkFilter(b *testing.B) {
    contracts := generateTestContracts(10000)
    filter := &DeltaFilter{MinDelta: 0.25, MaxDelta: 0.35}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = filter.Apply(contracts)
    }
}
```

Target benchmarks:
- Simple filters: < 1μs per contract
- Greeks filters: < 5μs per contract
- Complex filters: < 10μs per contract

## Best Practices

### 1. Filter Composition
```go
// Combine related filters for efficiency
type CompositeGreeksFilter struct {
    Delta GreeksFilter
    Gamma GreeksFilter
    Theta GreeksFilter
}

func (f *CompositeGreeksFilter) Apply(contracts []Contract) []Contract {
    // Single pass through contracts
    filtered := make([]Contract, 0, len(contracts)/2)
    
    for _, contract := range contracts {
        if f.matchesAll(contract) {
            filtered = append(filtered, contract)
        }
    }
    
    return filtered
}
```

### 2. Error Handling
```go
func (f *IVPercentileFilter) Apply(contracts []Contract) ([]Contract, error) {
    if f.cache == nil {
        return nil, errors.New("IV history cache not initialized")
    }
    
    if f.LookbackDays < 20 {
        return nil, errors.New("lookback period too short for reliable percentile")
    }
    
    // Continue with filtering...
}
```

### 3. Configuration Validation
```go
func (f *DeltaFilter) Validate() error {
    if f.MinDelta < 0 || f.MinDelta > 1 {
        return fmt.Errorf("invalid MinDelta: %f", f.MinDelta)
    }
    if f.MaxDelta < f.MinDelta {
        return errors.New("MaxDelta must be >= MinDelta")
    }
    return nil
}
```

## Integration Example

```go
// Complete filter chain for conservative credit spread strategy
func createConservativeCreditSpreadFilters() *FilterChain {
    return &FilterChain{
        Filters: []Filter{
            // Most selective first
            &LiquidityFilter{
                MinVolume:       50,
                MinOpenInterest: 100,
                MaxBidAskSpread: 0.10,
            },
            &DTEFilter{
                MinDays: 30,
                MaxDays: 60,
            },
            &DeltaFilter{
                MinDelta: 0.20,
                MaxDelta: 0.35,
            },
            &IVPercentileFilter{
                MinPercentile: 70,
                MaxPercentile: 100,
                LookbackDays:  252,
            },
            &GreeksFilter{
                MaxGamma: 0.05,
                MinTheta: -0.15,
                MaxVega:  0.50,
            },
        },
        CacheTTL: 5 * time.Minute,
    }
}
```