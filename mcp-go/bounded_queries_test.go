package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func querySession(t *testing.T, resources, placements any) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.RawQuery != "" {
			t.Error("unexpected backend request")
		}
		switch r.URL.Path {
		case "/api/resources":
			json.NewEncoder(w).Encode(resources)
		case "/api/placements":
			json.NewEncoder(w).Encode(placements)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(h.Close)
	b, err := newBackend(h.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	st, ct := mcp.NewInMemoryTransports()
	ss, err := server(b, b).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "query-test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, ctx
}
func queryFixtures() (any, any) {
	nodes := []resourceItem{}
	for i := 0; i < 120; i++ {
		category := "berries"
		if i%2 != 0 {
			category = "mushrooms"
		}
		nodes = append(nodes, resourceItem{ID: fmt.Sprint(i), Category: category, Class: "synthetic", Position: []float64{float64(i), 0, 99}})
	}
	placements := []placementItem{}
	for i := 0; i < 4; i++ {
		placements = append(placements, placementItem{Schema: "bellwright.foraging-placement-candidate.v1", ID: fmt.Sprint(i), Radius: 100, Current: placementSite{Position: []float64{999, 999, 0}, Count: 1}, Practical: placementSite{Position: []float64{float64(i), 0, 99}, Count: 3}, Assumptions: []string{"synthetic radius unverified"}})
	}
	return map[string]any{"schema": "bellwright.harvestable-baseline.v1", "source_kind": "packaged_game_asset", "nodes": nodes, "unknowns": []string{"synthetic availability unknown"}}, placements
}
func callQuery(t *testing.T, cs *mcp.ClientSession, ctx context.Context, name string, args any) map[string]any {
	t.Helper()
	result, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil || result.IsError {
		t.Fatalf("call %s: %+v %v", name, result, err)
	}
	raw, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err = json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range tools.Tools {
		if tool.Name == name {
			schemaRaw, err := json.Marshal(tool.OutputSchema)
			if err != nil {
				t.Fatal(err)
			}
			var schema jsonschema.Schema
			if err = json.Unmarshal(schemaRaw, &schema); err != nil {
				t.Fatal(err)
			}
			resolved, err := schema.Resolve(nil)
			if err != nil {
				t.Fatal(err)
			}
			if err = resolved.Validate(out); err != nil {
				t.Fatalf("actual structured result violates advertised schema: %v", err)
			}
			return out
		}
	}
	t.Fatal("tool missing")
	return nil
}
func counts(t *testing.T, out map[string]any, total, matched, returned int) {
	t.Helper()
	for key, want := range map[string]int{"total": total, "matched": matched, "returned": returned, "truncated": matched - returned} {
		if out[key] != float64(want) {
			t.Fatalf("%s: got %v want %d", key, out[key], want)
		}
	}
	if out["availability"] != "unknown" || out["retrieved_at"] == "" || out["source"] == "" {
		t.Fatalf("provenance missing: %+v", out)
	}
	if len(out["items"].([]any)) != returned {
		t.Fatal("item/count mismatch")
	}
}
func TestBoundedResourceMCP(t *testing.T) {
	resources, placements := queryFixtures()
	cs, ctx := querySession(t, resources, placements)
	tests := []struct {
		name              string
		args              map[string]any
		matched, returned int
	}{
		{"default", map[string]any{}, 120, 25},
		{"hard maximum", map[string]any{"limit": 100}, 120, 100},
		{"category", map[string]any{"category": "berries", "limit": 2}, 60, 2},
		{"radius inclusive XY", map[string]any{"center": map[string]any{"x": 0, "y": 0}, "radius": 2}, 3, 3},
		{"intersection", map[string]any{"category": "berries", "center": map[string]any{"x": 0, "y": 0}, "radius": 6, "bounds": map[string]any{"min_x": 2, "min_y": 0, "max_x": 4, "max_y": 0}}, 2, 2},
		{"no match", map[string]any{"category": "unobserved"}, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := callQuery(t, cs, ctx, "read_resource_baseline", tt.args)
			counts(t, out, 120, tt.matched, tt.returned)
		})
	}
}
func TestBoundedPlacementMCP(t *testing.T) {
	resources, placements := queryFixtures()
	cs, ctx := querySession(t, resources, placements)
	out := callQuery(t, cs, ctx, "read_foraging_candidates", map[string]any{"limit": 1})
	counts(t, out, 4, 4, 1)
	out = callQuery(t, cs, ctx, "read_foraging_candidates", map[string]any{"center": map[string]any{"x": 0, "y": 0}, "radius": 1})
	counts(t, out, 4, 2, 2)
	out = callQuery(t, cs, ctx, "read_foraging_candidates", map[string]any{"target_runtime_actor_id": "3"})
	counts(t, out, 4, 1, 1)
	item := out["items"].([]any)[0].(map[string]any)
	if len(item["assumptions"].([]any)) == 0 || item["current"] == nil || item["practical_candidate"] == nil {
		t.Fatal("lost placement evidence")
	}
}
func TestInvalidQueryMCP(t *testing.T) {
	resources, placements := queryFixtures()
	cs, ctx := querySession(t, resources, placements)
	invalid := []map[string]any{
		{"limit": 0}, {"limit": 101}, {"limit": -1}, {"limit": 1.5},
		{"radius": 1}, {"center": map[string]any{"x": 0, "y": 0}},
		{"center": map[string]any{"x": 0, "y": 0}, "radius": 0},
		{"center": map[string]any{"x": 0, "y": 0}, "radius": -1},
		{"center": map[string]any{"x": 1e9, "y": 0}, "radius": 1},
		{"center": map[string]any{"x": 0}, "radius": 1},
		{"bounds": map[string]any{"min_x": 2, "min_y": 0, "max_x": 1, "max_y": 1}},
		{"bounds": map[string]any{"min_x": 0}}, {"limit": nil}, {"arbitrary_url": "http://example.com"},
	}
	for _, name := range []string{"read_resource_baseline", "read_foraging_candidates"} {
		for i, args := range invalid {
			t.Run(fmt.Sprintf("%s/%d", name, i), func(t *testing.T) {
				r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
				if e == nil && !r.IsError {
					t.Fatalf("accepted invalid args %+v", args)
				}
			})
		}
	}
}
func TestMissingBackendPositionMCP(t *testing.T) {
	resources, placements := queryFixtures()
	resources.(map[string]any)["nodes"] = []map[string]any{{"id": "x", "class_asset_path": "synthetic", "baseline_category": "berries"}}
	cs, ctx := querySession(t, resources, placements)
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_resource_baseline", Arguments: map[string]any{}})
	if e == nil && !r.IsError {
		t.Fatal("missing coordinate evidence accepted")
	}
}
