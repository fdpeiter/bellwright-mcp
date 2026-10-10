package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func historySession(t *testing.T, data any) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/trends/item" || r.URL.Query().Get("name") != "Iron_C" || len(r.URL.Query()) != 1 {
			t.Error("unexpected item query")
		}
		json.NewEncoder(w).Encode(data)
	}))
	t.Cleanup(h.Close)
	b, _ := newBackend(h.URL)
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	registerItemHistoryTools(s, b)
	st, ct := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, st, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { ss.Close() })
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, ctx
}
func TestItemHistoryMCP(t *testing.T) {
	cs, ctx := historySession(t, map[string]any{"name": "Iron_C", "total_snapshots": 500, "sampled_points": 2, "stride": 3, "sampled": true, "points": []any{map[string]any{"ingested_at": "older", "qty": 0, "playtime": 0}, map[string]any{"snapshot_id": 9, "ingested_at": "newer", "observed_qty": 4, "observed_playtime": 20, "character": "Alice", "map": "World"}}})
	o := callQuery(t, cs, ctx, "read_item_history", map[string]any{"item_id": "Iron_C", "limit": 1})
	p := o["points"].([]any)[0].(map[string]any)
	if o["matched"] != float64(2) || o["truncated"] != float64(1) || p["ingested_at"] != "newer" || p["observed_qty"] != float64(4) || p["character"] != "Alice" {
		t.Fatal(o)
	}
	o = callQuery(t, cs, ctx, "read_item_history", map[string]any{"item_id": "Iron_C"})
	p = o["points"].([]any)[0].(map[string]any)
	if p["observed_qty"] != nil || p["observed_playtime"] != nil {
		t.Fatal("legacy zero became observed")
	}
	tools, _ := cs.ListTools(ctx, nil)
	if !tools.Tools[0].Annotations.ReadOnlyHint {
		t.Fatal("not readonly")
	}
	for _, a := range []any{map[string]any{}, map[string]any{"item_id": "../secret"}, map[string]any{"item_id": "Iron_C&name=x"}, map[string]any{"item_id": "Iron_C", "limit": 101}} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_item_history", Arguments: a})
		if e == nil && !r.IsError {
			t.Fatal("accepted invalid query", a)
		}
	}
}
func TestItemHistoryEmpty(t *testing.T) {
	cs, ctx := historySession(t, map[string]any{"name": "Iron_C", "points": []any{}})
	o := callQuery(t, cs, ctx, "read_item_history", map[string]any{"item_id": "Iron_C"})
	if o["returned"] != float64(0) || o["sampled"] != nil {
		t.Fatal(o)
	}
}
func TestItemHistoryMalformed(t *testing.T) {
	for _, d := range []any{nil, map[string]any{"name": "other", "points": []any{}}, map[string]any{"name": "Iron_C", "points": []any{map[string]any{"ingested_at": "x", "observed_qty": -1}}}} {
		cs, ctx := historySession(t, d)
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_item_history", Arguments: map[string]any{"item_id": "Iron_C"}})
		if e == nil && !r.IsError {
			t.Fatal("accepted malformed history")
		}
	}
}
