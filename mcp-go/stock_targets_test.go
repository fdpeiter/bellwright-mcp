package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
	"testing"
)

func stockTargetsFixture(t *testing.T, raw string, q stockTargetQuery) stockTargetOutput {
	t.Helper()
	b := healthFixture(t, map[string]string{"/api/world": raw})
	_, o, e := b.stockTargets(context.Background(), nil, q)
	if e != nil {
		t.Fatal(e)
	}
	return o
}
func TestStockTargetsScopeUnknownAndBounds(t *testing.T) {
	target := 5.0
	one := 1
	q := stockTargetQuery{Targets: map[string]*float64{"Iron": &target, "Absent": &target}, Scope: "stored", Limit: &one}
	o := stockTargetsFixture(t, `{"snapshot_id":3,"ingested_at":"old","storage":{"totals":{"Iron":2},"npc_carried":{"Iron":20}},"npcs":[{"is_player_npc":true,"equipment":{"hand":"Iron"}}]}`, q)
	if o.DeficitItems != 2 || o.Truncated != 1 || *o.Items[0].Quantity != 0 || *o.IngestedAt != "old" {
		t.Fatal(o)
	}
	q.Limit = nil
	q.Targets = map[string]*float64{"Iron": &target}
	q.Scope = "npc_carried"
	o = stockTargetsFixture(t, `{"storage":{"totals":{"Iron":2},"npc_carried":{"Iron":20}}}`, q)
	if *o.Items[0].Surplus != 15 {
		t.Fatal(o)
	}
	q.Scope = "equipped"
	o = stockTargetsFixture(t, `{"npcs":[{"is_player_npc":true,"equipment":{"hand":"Iron"}},{"is_player_npc":false,"equipment":{"hand":"Iron"}}]}`, q)
	if *o.Items[0].Quantity != 1 {
		t.Fatal(o)
	}
	for _, raw := range []string{`{}`, `{"npcs":null}`, `{"npcs":[{"equipment":{}}]}`, `{"npcs":[{"is_player_npc":true}]}`, `{"npcs":[{"is_player_npc":true,"equipment":{"hand":null}}]}`} {
		o = stockTargetsFixture(t, raw, q)
		if o.Unknown != 1 || o.Items[0].Quantity != nil || o.Items[0].Deficit != nil {
			t.Fatalf("%s: %+v", raw, o)
		}
	}
	q.Scope = "stored"
	for _, raw := range []string{`{}`, `{"storage":{"totals":null}}`, `{"storage":{"totals":{"Iron":null}}}`, `{"storage":{"totals":{"Iron":"2"}}}`, `{"storage":{"totals":{"Iron":-1}}}`, `{"storage":{"totals":{"Iron":1.5}}}`} {
		o = stockTargetsFixture(t, raw, q)
		if o.Unknown != 1 || o.Items[0].Quantity != nil {
			t.Fatalf("%s: %+v", raw, o)
		}
	}
	o = stockTargetsFixture(t, `{"storage":{"totals":{"Other":null}}}`, q)
	if o.Unknown != 1 {
		t.Fatal("malformed aggregate assumed complete")
	}
	o = stockTargetsFixture(t, `{"storage":{"totals":{"Iron":0}}}`, q)
	if o.Unknown != 0 || *o.Items[0].Quantity != 0 {
		t.Fatal(o)
	}
}
func TestStockTargetFailures(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{}`})
	for _, q := range []stockTargetQuery{{Scope: "stored"}, {Scope: "all", Targets: map[string]*float64{"x": floatPtr(1)}}, {Scope: "stored", Targets: map[string]*float64{"x": nil}}, {Scope: "stored", Targets: map[string]*float64{"x": floatPtr(-1)}}, {Scope: "stored", Targets: map[string]*float64{"x": floatPtr(1.5)}}, {Scope: "stored", Targets: map[string]*float64{"x": floatPtr(math.Inf(1))}}} {
		_, _, e := b.stockTargets(context.Background(), nil, q)
		if e == nil {
			t.Fatal("accepted invalid target", q)
		}
	}
}
func floatPtr(v float64) *float64 { return &v }
func TestStockTargetsMCP(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"storage":{"totals":{"Iron":2}}}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "stock-test", Version: "1"}, nil)
	registerStockTargetTools(s, b)
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
	o := callQuery(t, cs, ctx, "compare_saved_stock_targets", map[string]any{"targets": map[string]any{"Iron": 5}, "scope": "stored"})
	if o["deficit_items"] != float64(1) {
		t.Fatal(o)
	}
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "compare_saved_stock_targets", Arguments: map[string]any{"targets": map[string]any{"Iron": nil}, "scope": "stored"}})
	if e == nil && !r.IsError {
		t.Fatal("accepted null target")
	}
}
