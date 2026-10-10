package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func gearSession(t *testing.T, data any) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gear" || r.URL.RawQuery != "" {
			t.Error("unexpected request")
		}
		json.NewEncoder(w).Encode(data)
	}))
	t.Cleanup(h.Close)
	b, _ := newBackend(h.URL)
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	registerGearTools(s, b)
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
func TestGearMCP(t *testing.T) {
	cs, ctx := gearSession(t, map[string]any{"snapshot_id": 7, "ingested_at": "save-time", "craft": []gearCraft{{Item: "Sword", Deficit: 2}, {Item: "Shield", Deficit: 1}}, "reequip": []any{}, "presets": []any{}, "kits_file": "private"})
	o := callQuery(t, cs, ctx, "read_gear_plan", map[string]any{"limit": 1})
	c := o["craft"].(map[string]any)
	if c["matched"] != float64(2) || c["returned"] != float64(1) || c["truncated"] != float64(1) || o["ingested_at"] != "save-time" {
		t.Fatal(o)
	}
	if _, ok := o["kits_file"]; ok {
		t.Fatal("private path leaked")
	}
	tools, _ := cs.ListTools(ctx, nil)
	if !tools.Tools[0].Annotations.ReadOnlyHint {
		t.Fatal("not readonly")
	}
	for _, a := range []any{map[string]any{"limit": 0}, map[string]any{"limit": 101}, map[string]any{"section": "bad"}, map[string]any{"kits_file": "bad"}} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_gear_plan", Arguments: a})
		if e == nil && !r.IsError {
			t.Fatal("accepted invalid input", a)
		}
	}
}
func TestGearEmpty(t *testing.T) {
	cs, ctx := gearSession(t, map[string]any{"craft": []any{}, "reequip": []any{}, "presets": []any{}})
	o := callQuery(t, cs, ctx, "read_gear_plan", map[string]any{})
	if o["snapshot_id"] != nil || o["ingested_at"] != nil {
		t.Fatal(o)
	}
}
func TestGearMalformed(t *testing.T) {
	for _, d := range []any{nil, map[string]any{}, map[string]any{"craft": "bad", "reequip": []any{}, "presets": []any{}}} {
		cs, ctx := gearSession(t, d)
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_gear_plan", Arguments: map[string]any{}})
		if e == nil && !r.IsError {
			t.Fatal("accepted malformed evidence")
		}
	}
}
