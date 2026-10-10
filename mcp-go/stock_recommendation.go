package main

import (
	"context"
	"fmt"
	"math"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type stockRecommendationPlan struct {
	Item                 string   `json:"item"`
	RatePerHour          *float64 `json:"rate_per_hour"`
	ReplenishmentHours   *float64 `json:"replenishment_hours"`
	SafetyBufferFraction *float64 `json:"safety_buffer_fraction"`
	PlannedDemand        *float64 `json:"planned_demand"`
}
type stockRecommendationQuery struct {
	Plans []stockRecommendationPlan `json:"plans"`
	Limit *int                      `json:"limit,omitempty"`
}
type stockRecommendationItem struct {
	Item                      string                  `json:"item"`
	UserAssumptions           stockRecommendationPlan `json:"user_assumptions"`
	DemandDuringReplenishment float64                 `json:"demand_during_replenishment"`
	SafetyStock               float64                 `json:"safety_stock"`
	PlannedDemand             float64                 `json:"planned_demand"`
	Target                    float64                 `json:"target"`
	Current                   *float64                `json:"current"`
	Deficit                   *float64                `json:"deficit"`
	Surplus                   *float64                `json:"surplus"`
	State                     string                  `json:"state"`
}
type stockRecommendationOutput struct {
	Source       string                    `json:"source"`
	SnapshotID   *int64                    `json:"snapshot_id"`
	IngestedAt   *string                   `json:"ingested_at"`
	RetrievedAt  string                    `json:"retrieved_at"`
	Scope        string                    `json:"scope"`
	State        string                    `json:"state"`
	Total        int                       `json:"total"`
	Returned     int                       `json:"returned"`
	Truncated    int                       `json:"truncated"`
	Limit        int                       `json:"limit"`
	Unknown      int                       `json:"unknown"`
	DeficitItems int                       `json:"deficit_items"`
	Items        []stockRecommendationItem `json:"items"`
	Formula      string                    `json:"formula"`
	Note         string                    `json:"note"`
}

var stockRecommendationItemKey = regexp.MustCompile(`^[A-Za-z0-9_]{1,256}$`)

const safeStockInteger = 9007199254740991.0

func registerStockRecommendationTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "recommend_stock_targets", Description: "Recommend stored stock targets from 1 through 100 explicit user assumptions. Each plan requires an exact ASCII item key, nonnegative rate_per_hour, positive replenishment_hours, safety_buffer_fraction between 0 and 2, and nonnegative integer planned_demand. Service-buffer sizing only; not measured consumption or global optimization. Read-only. Limit defaults 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.stockRecommendations)
}
func (b *backend) stockRecommendations(ctx context.Context, req *mcp.CallToolRequest, q stockRecommendationQuery) (*mcp.CallToolResult, stockRecommendationOutput, error) {
	o := stockRecommendationOutput{Source: "companion_save", Scope: "stored", State: "unknown", Items: []stockRecommendationItem{}, Formula: "ceil(rate_per_hour * replenishment_hours * (1 + safety_buffer_fraction) + planned_demand)", Note: "Sizing policy from explicit user assumptions, not measured consumption or a globally optimal allocation. No history-derived rates, recipes, capacity constraints, production scheduling or persistent changes. Current quantities come from the latest ingested saved stored aggregate; ingestion and retrieval are not capture times. Missing keys count as zero only in an observed complete aggregate; null/malformed quantities remain unknown."}
	if len(q.Plans) < 1 || len(q.Plans) > 100 {
		return nil, o, fmt.Errorf("require 1 through 100 plans")
	}
	limit := 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > 100 {
		return nil, o, fmt.Errorf("limit must be 1 through 100")
	}
	targets := map[string]*float64{}
	rows := map[string]stockRecommendationItem{}
	finiteSafe := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= safeStockInteger }
	for _, p := range q.Plans {
		if !stockRecommendationItemKey.MatchString(p.Item) {
			return nil, o, fmt.Errorf("item must be an exact ASCII alphanumeric or underscore key of 1 through 256 characters")
		}
		if _, exists := targets[p.Item]; exists {
			return nil, o, fmt.Errorf("duplicate item %s", p.Item)
		}
		if p.RatePerHour == nil || p.ReplenishmentHours == nil || p.SafetyBufferFraction == nil || p.PlannedDemand == nil {
			return nil, o, fmt.Errorf("all assumptions are required and must be nonnull")
		}
		if !finiteSafe(*p.RatePerHour) || !finiteSafe(*p.ReplenishmentHours) || *p.ReplenishmentHours <= 0 || !finiteSafe(*p.SafetyBufferFraction) || *p.SafetyBufferFraction > 2 || !finiteSafe(*p.PlannedDemand) || math.Trunc(*p.PlannedDemand) != *p.PlannedDemand {
			return nil, o, fmt.Errorf("invalid or unsafe stock assumptions")
		}
		demand := *p.RatePerHour * *p.ReplenishmentHours
		safety := demand * *p.SafetyBufferFraction
		buffered := demand * (1 + *p.SafetyBufferFraction)
		sum := buffered + *p.PlannedDemand
		target := math.Ceil(sum)
		if !finiteSafe(demand) || !finiteSafe(safety) || !finiteSafe(buffered) || !finiteSafe(sum) || !finiteSafe(target) {
			return nil, o, fmt.Errorf("calculated stock target or intermediate exceeds safe numeric bounds")
		}
		targets[p.Item] = &target
		rows[p.Item] = stockRecommendationItem{Item: p.Item, UserAssumptions: p, DemandDuringReplenishment: demand, SafetyStock: safety, PlannedDemand: *p.PlannedDemand, Target: target}
	}
	_, comparison, err := b.stockTargets(ctx, req, stockTargetQuery{Targets: targets, Scope: "stored", Limit: &limit})
	if err != nil {
		return nil, o, err
	}
	o.SnapshotID = comparison.SnapshotID
	o.IngestedAt = comparison.IngestedAt
	o.RetrievedAt = comparison.RetrievedAt
	o.State = comparison.State
	o.Total = comparison.Total
	o.Returned = comparison.Returned
	o.Truncated = comparison.Truncated
	o.Limit = comparison.Limit
	o.Unknown = comparison.Unknown
	o.DeficitItems = comparison.DeficitItems
	for _, r := range comparison.Items {
		row := rows[r.Item]
		row.Current = r.Quantity
		row.Deficit = r.Deficit
		row.Surplus = r.Surplus
		row.State = r.State
		o.Items = append(o.Items, row)
	}
	return nil, o, nil
}
