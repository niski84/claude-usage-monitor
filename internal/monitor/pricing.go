package monitor

import "strings"

type modelPricing struct {
	InPerM         float64
	OutPerM        float64
	CacheWritePerM float64
	CacheReadPerM  float64
}

var pricing = map[string]modelPricing{
	"claude-opus-4-6":   {15.00, 75.00, 18.75, 1.50},
	"claude-opus-4-5":   {15.00, 75.00, 18.75, 1.50},
	"claude-opus-3":     {15.00, 75.00, 18.75, 1.50},
	"claude-sonnet-4-6": {3.00, 15.00, 3.75, 0.30},
	"claude-sonnet-4-5": {3.00, 15.00, 3.75, 0.30},
	"claude-3-5-sonnet": {3.00, 15.00, 3.75, 0.30},
	"claude-haiku-4-5":  {0.80, 4.00, 1.00, 0.08},
	"claude-3-5-haiku":  {0.80, 4.00, 1.00, 0.08},
	"claude-haiku-3":    {0.25, 1.25, 0.30, 0.03},
}

var defaultPricing = modelPricing{3.00, 15.00, 3.75, 0.30}

func pricingFor(model string) modelPricing {
	// exact match
	if p, ok := pricing[model]; ok {
		return p
	}
	// prefix match (handles minor version suffixes)
	for k, p := range pricing {
		if strings.HasPrefix(model, k) || strings.HasPrefix(k, model) {
			return p
		}
	}
	return defaultPricing
}

// CostForRecord computes the USD cost for a single usage record.
func CostForRecord(r *UsageRecord) float64 {
	p := pricingFor(r.Model)
	cost := float64(r.InputTokens)*p.InPerM/1e6 +
		float64(r.OutputTokens)*p.OutPerM/1e6 +
		float64(r.CacheWriteTokens)*p.CacheWritePerM/1e6 +
		float64(r.CacheReadTokens)*p.CacheReadPerM/1e6
	return cost
}
