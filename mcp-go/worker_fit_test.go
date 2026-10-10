package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func workerFitSession(t *testing.T, world any, register func(*mcp.Server, *backend)) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/world" || r.Method != "GET" || r.URL.RawQuery != "" {
			t.Error("unexpected request")
		}
		json.NewEncoder(w).Encode(world)
	}))
	t.Cleanup(h.Close)
	b, e := newBackend(h.URL)
	if e != nil {
		t.Fatal(e)
	}
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "fit", Version: "1"}, nil)
	register(s, b)
	st, ct := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, st, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { ss.Close() })
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "fit", Version: "1"}, nil).Connect(ctx, ct, nil)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, ctx
}
func TestWorkerFitMCP(t *testing.T) {
	cs, ctx := workerFitSession(t, map[string]any{"snapshot_id": 4, "ingested_at": "old", "npcs": []any{map[string]any{"guid": "a", "is_player_npc": true, "skills": map[string]any{"farm": map[string]any{"level": 3}}, "injuries": nil, "traits": []any{}}, map[string]any{"guid": "b", "is_player_npc": true, "skills": map[string]any{"farm": map[string]any{"level": 3}}}, map[string]any{"is_player_npc": true}, map[string]any{"is_player_npc": true, "skills": "bad"}, map[string]any{"is_player_npc": false}}}, registerWorkerFitTools)
	o := callQuery(t, cs, ctx, "rank_saved_workers", map[string]any{"skill": "farm", "limit": 1})
	if o["matched"] != float64(2) || o["excluded_unknown_skill"] != float64(1) || o["excluded_malformed_skill"] != float64(1) || o["truncated"] != float64(1) {
		t.Fatal(o)
	}
	i := o["items"].([]any)[0].(map[string]any)
	if i["rank"] != float64(1) || i["tied_workers"] != float64(2) || i["observations"].(map[string]any)["guid"] != "a" {
		t.Fatal(i)
	}
	for _, args := range []any{map[string]any{}, map[string]any{"skill": "hunting"}, map[string]any{"skill": "farm", "limit": 0}, map[string]any{"skill": "farm", "limit": 101}} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "rank_saved_workers", Arguments: args})
		if e == nil && !r.IsError {
			t.Fatal(args)
		}
	}
}
func TestWorkerFitMissingMalformed(t *testing.T) {
	for _, world := range []any{map[string]any{}, map[string]any{"npcs": "bad"}} {
		cs, ctx := workerFitSession(t, world, registerWorkerFitTools)
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "rank_saved_workers", Arguments: map[string]any{"skill": "farm"}})
		if e == nil && !r.IsError {
			t.Fatal(world)
		}
	}
	cs, ctx := workerFitSession(t, map[string]any{"npcs": []any{}}, registerWorkerFitTools)
	o := callQuery(t, cs, ctx, "rank_saved_workers", map[string]any{"skill": "farm", "limit": 100})
	if o["returned"] != float64(0) || o["snapshot_id"] != nil {
		t.Fatal(o)
	}
}
