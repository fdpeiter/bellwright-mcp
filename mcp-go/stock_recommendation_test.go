package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
	"testing"
)

func recommendationPlan(item string) stockRecommendationPlan {
	return stockRecommendationPlan{Item: item, RatePerHour: floatPtr(1.5), ReplenishmentHours: floatPtr(3), SafetyBufferFraction: floatPtr(.25), PlannedDemand: floatPtr(2)}
}
func TestStockRecommendationSizing(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"snapshot_id":4,"storage":{"totals":{"Iron_C":3,"Unknown_C":null},"npc_carried":{"Iron_C":99}}}`})
	q := stockRecommendationQuery{Plans: []stockRecommendationPlan{recommendationPlan("Unknown_C"), recommendationPlan("Iron_C"), recommendationPlan("Absent_C")}}
	_, o, e := b.stockRecommendations(context.Background(), nil, q)
	if e != nil {
		t.Fatal(e)
	}
	if o.Scope != "stored" || o.Total != 3 || o.Unknown != 2 || o.DeficitItems != 1 || *o.SnapshotID != 4 {
		t.Fatal(o)
	}
	r := o.Items[1]
	if r.Item != "Iron_C" || r.Target != 8 || r.DemandDuringReplenishment != 4.5 || r.SafetyStock != 1.125 || r.PlannedDemand != 2 || *r.Current != 3 || *r.Deficit != 5 || *r.Surplus != 0 {
		t.Fatal(r)
	}
	if o.Items[0].Current != nil || o.Items[2].Current != nil || o.Items[2].Deficit != nil {
		t.Fatal(o)
	}
	one := 1
	q.Limit = &one
	_, o, e = b.stockRecommendations(context.Background(), nil, q)
	if e != nil || o.Returned != 1 || o.Truncated != 2 || o.Unknown != 2 {
		t.Fatal(o, e)
	}
	b = healthFixture(t, map[string]string{"/api/world": `{"storage":{"totals":{}}}`})
	_, o, e = b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: []stockRecommendationPlan{recommendationPlan("Absent_C")}})
	if e != nil || *o.Items[0].Current != 0 || *o.Items[0].Deficit != 8 {
		t.Fatal(o, e)
	}
}
func TestStockRecommendationValidation(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{}`})
	bad := []stockRecommendationPlan{}
	for _, f := range []string{"rate", "hours", "buffer", "planned"} {
		p := recommendationPlan("Iron_C")
		switch f {
		case "rate":
			p.RatePerHour = nil
		case "hours":
			p.ReplenishmentHours = nil
		case "buffer":
			p.SafetyBufferFraction = nil
		case "planned":
			p.PlannedDemand = nil
		}
		bad = append(bad, p)
	}
	for _, s := range []string{"", " Iron_C", "Iron-C", "Íron_C"} {
		bad = append(bad, recommendationPlan(s))
	}
	for _, v := range []float64{-1, math.NaN(), math.Inf(1), safeStockInteger + 1} {
		p := recommendationPlan("Iron_C")
		p.RatePerHour = floatPtr(v)
		bad = append(bad, p)
	}
	p := recommendationPlan("Iron_C")
	p.ReplenishmentHours = floatPtr(0)
	bad = append(bad, p)
	p = recommendationPlan("Iron_C")
	p.SafetyBufferFraction = floatPtr(2.01)
	bad = append(bad, p)
	p = recommendationPlan("Iron_C")
	p.PlannedDemand = floatPtr(.5)
	bad = append(bad, p)
	p = recommendationPlan("Iron_C")
	p.RatePerHour = floatPtr(safeStockInteger)
	p.ReplenishmentHours = floatPtr(2)
	bad = append(bad, p)
	p = recommendationPlan("Iron_C")
	p.RatePerHour = floatPtr(safeStockInteger / 2)
	p.ReplenishmentHours = floatPtr(1)
	p.SafetyBufferFraction = floatPtr(2)
	bad = append(bad, p)
	p = recommendationPlan("Iron_C")
	p.RatePerHour = floatPtr(safeStockInteger)
	p.ReplenishmentHours = floatPtr(1)
	p.SafetyBufferFraction = floatPtr(0)
	p.PlannedDemand = floatPtr(2)
	bad = append(bad, p)
	for _, p := range bad {
		_, _, e := b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: []stockRecommendationPlan{p}})
		if e == nil {
			t.Fatalf("accepted %+v", p)
		}
	}
	_, _, e := b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: []stockRecommendationPlan{recommendationPlan("Iron_C"), recommendationPlan("Iron_C")}})
	if e == nil {
		t.Fatal("accepted duplicate")
	}
}
func TestStockRecommendationMCPSchema(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"storage":{"totals":{"Iron_C":3}}}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	registerStockRecommendationTools(s, b)
	st, ct := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, st, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer ss.Close()
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer cs.Close()
	p := map[string]any{"item": "Iron_C", "rate_per_hour": 1.5, "replenishment_hours": 3, "safety_buffer_fraction": .25, "planned_demand": 2}
	o := callQuery(t, cs, ctx, "recommend_stock_targets", map[string]any{"plans": []any{p}})
	if o["scope"] != "stored" || o["deficit_items"] != float64(1) {
		t.Fatal(o)
	}
	for _, field := range []string{"rate_per_hour", "replenishment_hours", "safety_buffer_fraction", "planned_demand"} {
		for _, missing := range []bool{false, true} {
			bad := map[string]any{}
			for k, v := range p {
				bad[k] = v
			}
			if missing {
				delete(bad, field)
			} else {
				bad[field] = nil
			}
			r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "recommend_stock_targets", Arguments: map[string]any{"plans": []any{bad}}})
			if e == nil && !r.IsError {
				t.Fatalf("accepted missing/null %s", field)
			}
		}
	}
}
func TestStockRecommendationBoundsAndZero(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"storage":{"totals":{}}}`})
	for _, limit := range []int{0, 101} {
		_, _, e := b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: []stockRecommendationPlan{recommendationPlan("Iron_C")}, Limit: &limit})
		if e == nil {
			t.Fatal("accepted invalid limit")
		}
	}
	for _, n := range []int{0, 101} {
		plans := make([]stockRecommendationPlan, n)
		for i := range plans {
			plans[i] = recommendationPlan("Iron_C")
		}
		_, _, e := b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: plans})
		if e == nil {
			t.Fatal("accepted invalid plan count")
		}
	}
	p := recommendationPlan("Zero_C")
	p.RatePerHour = floatPtr(0)
	p.PlannedDemand = floatPtr(0)
	_, o, e := b.stockRecommendations(context.Background(), nil, stockRecommendationQuery{Plans: []stockRecommendationPlan{p}})
	if e != nil || o.Items[0].Target != 0 || *o.Items[0].Deficit != 0 {
		t.Fatal(o, e)
	}
}
