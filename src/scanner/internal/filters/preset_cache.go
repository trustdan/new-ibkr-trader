package filters

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PresetCache manages filter presets
type PresetCache struct {
	mu      sync.RWMutex
	presets map[string]*FilterPreset
}

// FilterPreset represents a saved filter configuration
type FilterPreset struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Filters     FilterConfig `json:"filters"`
	Tags        []string     `json:"tags"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	UsageCount  int          `json:"usage_count"`
}

// NewPresetCache creates a new preset cache
func NewPresetCache() *PresetCache {
	cache := &PresetCache{
		presets: make(map[string]*FilterPreset),
	}
	
	// Load default presets
	cache.loadDefaults()
	
	return cache
}

// loadDefaults loads default filter presets
func (pc *PresetCache) loadDefaults() {
	defaults := []struct {
		name        string
		description string
		filters     FilterConfig
		tags        []string
	}{
		{
			name:        "Conservative",
			description: "Low risk, steady returns strategy",
			filters: FilterConfig{
				Delta: &DeltaFilter{
					Min: 0.25,
					Max: 0.35,
				},
				DTE: &DTEFilter{
					Min: 30,
					Max: 60,
				},
				Liquidity: &LiquidityFilter{
					MinOpenInterest: 500,
					MinVolume:       100,
				},
				Spread: &SpreadFilter{
					MinCredit:    1.0,
					MaxWidth:     5.0,
					MinRiskReward: 0.5,
				},
			},
			tags: []string{"conservative", "low-risk", "beginner"},
		},
		{
			name:        "Balanced",
			description: "Balanced risk/reward strategy",
			filters: FilterConfig{
				Delta: &DeltaFilter{
					Min: 0.20,
					Max: 0.30,
				},
				DTE: &DTEFilter{
					Min: 21,
					Max: 45,
				},
				Liquidity: &LiquidityFilter{
					MinOpenInterest: 250,
					MinVolume:       50,
				},
				Spread: &SpreadFilter{
					MinCredit:    0.75,
					MaxWidth:     7.5,
					MinRiskReward: 0.4,
				},
			},
			tags: []string{"balanced", "moderate", "popular"},
		},
		{
			name:        "Aggressive",
			description: "Higher risk, higher potential returns",
			filters: FilterConfig{
				Delta: &DeltaFilter{
					Min: 0.15,
					Max: 0.25,
				},
				DTE: &DTEFilter{
					Min: 14,
					Max: 30,
				},
				Liquidity: &LiquidityFilter{
					MinOpenInterest: 100,
					MinVolume:       25,
				},
				Spread: &SpreadFilter{
					MinCredit:    0.50,
					MaxWidth:     10.0,
					MinRiskReward: 0.3,
				},
			},
			tags: []string{"aggressive", "high-risk", "experienced"},
		},
		{
			name:        "Weekly Income",
			description: "Short-term weekly strategies",
			filters: FilterConfig{
				Delta: &DeltaFilter{
					Min: 0.20,
					Max: 0.30,
				},
				DTE: &DTEFilter{
					Min: 5,
					Max: 10,
				},
				Liquidity: &LiquidityFilter{
					MinOpenInterest: 1000,
					MinVolume:       200,
				},
				Spread: &SpreadFilter{
					MinCredit:    0.25,
					MaxWidth:     2.5,
					MinRiskReward: 0.3,
				},
			},
			tags: []string{"weekly", "income", "short-term"},
		},
		{
			name:        "High Probability",
			description: "Focus on high probability of profit",
			filters: FilterConfig{
				Delta: &DeltaFilter{
					Min: 0.10,
					Max: 0.20,
				},
				DTE: &DTEFilter{
					Min: 30,
					Max: 45,
				},
				Liquidity: &LiquidityFilter{
					MinOpenInterest: 500,
					MinVolume:       100,
				},
				Spread: &SpreadFilter{
					MinCredit:    0.50,
					MaxWidth:     5.0,
					MinRiskReward: 0.25,
				},
				Advanced: &AdvancedFilter{
					MinPoP: 0.70,
				},
			},
			tags: []string{"high-pop", "probability", "safe"},
		},
	}
	
	for _, preset := range defaults {
		id := uuid.New().String()
		pc.presets[id] = &FilterPreset{
			ID:          id,
			Name:        preset.name,
			Description: preset.description,
			Filters:     preset.filters,
			Tags:        preset.tags,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			UsageCount:  0,
		}
	}
}

// GetAll returns all presets
func (pc *PresetCache) GetAll() map[string]*FilterPreset {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make(map[string]*FilterPreset, len(pc.presets))
	for k, v := range pc.presets {
		result[k] = v
	}
	
	return result
}

// Get returns a specific preset
func (pc *PresetCache) Get(id string) (*FilterPreset, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	preset, exists := pc.presets[id]
	if exists {
		// Increment usage count
		go pc.incrementUsage(id)
	}
	
	return preset, exists
}

// Save creates a new preset
func (pc *PresetCache) Save(name, description string, filters FilterConfig, tags []string) string {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	id := uuid.New().String()
	pc.presets[id] = &FilterPreset{
		ID:          id,
		Name:        name,
		Description: description,
		Filters:     filters,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UsageCount:  0,
	}
	
	return id
}

// Update updates an existing preset
func (pc *PresetCache) Update(id string, update interface{}) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	preset, exists := pc.presets[id]
	if !exists {
		return fmt.Errorf("preset not found: %s", id)
	}
	
	// Update fields based on provided data
	data, err := json.Marshal(update)
	if err != nil {
		return err
	}
	
	var updates struct {
		Name        string       `json:"name,omitempty"`
		Description string       `json:"description,omitempty"`
		Filters     FilterConfig `json:"filters,omitempty"`
		Tags        []string     `json:"tags,omitempty"`
	}
	
	if err := json.Unmarshal(data, &updates); err != nil {
		return err
	}
	
	// Apply updates
	if updates.Name != "" {
		preset.Name = updates.Name
	}
	if updates.Description != "" {
		preset.Description = updates.Description
	}
	if updates.Filters.Delta != nil || updates.Filters.DTE != nil {
		preset.Filters = updates.Filters
	}
	if updates.Tags != nil {
		preset.Tags = updates.Tags
	}
	
	preset.UpdatedAt = time.Now()
	
	return nil
}

// Delete removes a preset
func (pc *PresetCache) Delete(id string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	if _, exists := pc.presets[id]; !exists {
		return fmt.Errorf("preset not found: %s", id)
	}
	
	delete(pc.presets, id)
	return nil
}

// incrementUsage increments the usage count for a preset
func (pc *PresetCache) incrementUsage(id string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	
	if preset, exists := pc.presets[id]; exists {
		preset.UsageCount++
	}
}

// GetPopular returns the most used presets
func (pc *PresetCache) GetPopular(limit int) []*FilterPreset {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	// Convert to slice for sorting
	presets := make([]*FilterPreset, 0, len(pc.presets))
	for _, preset := range pc.presets {
		presets = append(presets, preset)
	}
	
	// Sort by usage count
	for i := 0; i < len(presets)-1; i++ {
		for j := i + 1; j < len(presets); j++ {
			if presets[i].UsageCount < presets[j].UsageCount {
				presets[i], presets[j] = presets[j], presets[i]
			}
		}
	}
	
	// Return top N
	if limit > len(presets) {
		limit = len(presets)
	}
	
	return presets[:limit]
}

// FindByTags returns presets matching any of the given tags
func (pc *PresetCache) FindByTags(tags []string) []*FilterPreset {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	
	result := make([]*FilterPreset, 0)
	
	for _, preset := range pc.presets {
		for _, tag := range tags {
			for _, presetTag := range preset.Tags {
				if tag == presetTag {
					result = append(result, preset)
					goto next
				}
			}
		}
	next:
	}
	
	return result
}