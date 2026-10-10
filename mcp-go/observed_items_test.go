package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func observedSession(t *testing.T, evidence any) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/world" || r.URL.RawQuery != "" || r.Method != "GET" {
			t.Error("unexpected request")
		}
		json.NewEncoder(w).Encode(evidence)
	}))
	t.Cleanup(h.Close)
	b, err := newBackend(h.URL)
	if err != nil {
		t.Fatal(err)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "observed-test", Version: "1"}, nil)
	registerObservedItemTools(s, b)
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "observed-client", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, ctx
}
func observedFixture() map[string]any {
	return map[string]any{
		"snapshot_id": 9, "ingested_at": "saved-time",
		"storage":      map[string]any{"totals": map[string]any{"SturdyWarhammer_C": 3, "Zero_C": 0}, "npc_carried": map[string]any{"Berry_C": 2}, "containers": []any{map[string]any{"items": map[string]any{"SturdyWarhammer_C": 3}}}},
		"npcs":         []any{map[string]any{"equipment": map[string]any{"weapon": "SturdyWarhammer_C", "empty": ""}}},
		"player":       map[string]any{"equipment": map[string]any{"shield": "Shield_C"}, "carried": []any{map[string]any{"item": "Apple_C", "qty": 4}}},
		"carried":      map[string]any{"actor": []any{map[string]any{"item": "Berry_C", "qty": 2}}},
		"gear_presets": []any{map[string]any{"slots": map[string]any{"weapon": []any{map[string]any{"item": "Unobserved_C", "rank": 1}}}}},
	}
}
func TestObservedItemsMCP(t *testing.T) {
	cs, ctx := observedSession(t, observedFixture())
	out := callQuery(t, cs, ctx, "search_observed_items", map[string]any{"limit": 2})
	if out["total"] != float64(4) || out["matched"] != float64(4) || out["returned"] != float64(2) || out["truncated"] != float64(2) || out["snapshot_id"] != float64(9) {
		t.Fatal(out)
	}
	items := out["items"].([]any)
	if items[0].(map[string]any)["item_class"] != "Apple_C" {
		t.Fatal(items)
	}
	out = callQuery(t, cs, ctx, "search_observed_items", map[string]any{"query": "STURDY WARHAMMER"})
	item := out["items"].([]any)[0].(map[string]any)
	if out["returned"] != float64(1) || item["official_label"] != nil || item["label_source"] != "identifier_format" || len(item["sources"].([]any)) != 3 {
		t.Fatal(out)
	}
	out = callQuery(t, cs, ctx, "search_observed_items", map[string]any{"query": "Unobserved"})
	if out["returned"] != float64(0) {
		t.Fatal("preset reference treated as observed")
	}
	out = callQuery(t, cs, ctx, "search_observed_items", map[string]any{"query": "Zero"})
	if out["returned"] != float64(0) {
		t.Fatal("zero quantity treated as observed")
	}
}
func TestObservedEmptyUnknown(t *testing.T) {
	cs, ctx := observedSession(t, map[string]any{"npcs": []any{}})
	out := callQuery(t, cs, ctx, "search_observed_items", map[string]any{})
	if out["total"] != float64(0) || out["snapshot_id"] != nil || out["ingested_at"] != nil || len(out["unknown_scopes"].([]any)) != 6 {
		t.Fatal(out)
	}
}
func TestObservedMalformed(t *testing.T) {
	cases := []any{nil, map[string]any{}, map[string]any{"npcs": "bad"}, map[string]any{"storage": map[string]any{"totals": map[string]any{"x": nil}}}, map[string]any{"storage": map[string]any{"totals": map[string]any{"x": -1}}}, map[string]any{"player": map[string]any{"carried": []any{map[string]any{"item": "x"}}}}, map[string]any{"npcs": []any{map[string]any{"equipment": map[string]any{"slot": nil}}}}, map[string]any{"carried": map[string]any{"actor": []any{map[string]any{"item": "", "qty": 1}}}}}
	for i, e := range cases {
		cs, ctx := observedSession(t, e)
		r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_observed_items", Arguments: map[string]any{}})
		if err == nil && !r.IsError {
			t.Fatalf("accepted case %d", i)
		}
	}
}
func TestObservedLimitsAndDown(t *testing.T) {
	cs, ctx := observedSession(t, observedFixture())
	for _, limit := range []int{0, 101} {
		r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_observed_items", Arguments: map[string]any{"limit": limit}})
		if err == nil && !r.IsError {
			t.Fatal("accepted bad limit")
		}
	}
	h := httptest.NewServer(http.NotFoundHandler())
	b, _ := newBackend(h.URL)
	h.Close()
	if _, _, err := b.observedItems(context.Background(), nil, observedItemQuery{}); err == nil {
		t.Fatal("accepted unavailable backend")
	}
}
func TestObservedFormatting(t *testing.T) {
	for id, want := range map[string]string{"SturdyWarhammer_C": "Sturdy Warhammer", "NPCSword_2_C": "NPC Sword", "Iron_Ore_C": "Iron Ore"} {
		if got := observedDisplayIdentifier(id); got != want {
			t.Fatalf("%q => %q", id, got)
		}
	}
}
