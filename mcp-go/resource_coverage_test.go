package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func coverageSession(t *testing.T, evidence any) (*mcp.ClientSession, context.Context) {
	t.Helper()
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/resources" || r.URL.RawQuery != "" || r.Method != "GET" {
			t.Error("unexpected request")
		}
		json.NewEncoder(w).Encode(evidence)
	}))
	t.Cleanup(h.Close)
	b, err := newBackend(h.URL)
	if err != nil {
		t.Fatal(err)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "coverage-test", Version: "1"}, nil)
	registerResourceCoverageTools(s, b)
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ss.Close() })
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "coverage-client", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs, ctx
}
func TestResourceCoverageMCP(t *testing.T) {
	resources, _ := queryFixtures()
	cs, ctx := coverageSession(t, resources)
	out := callQuery(t, cs, ctx, "summarize_resource_coverage", map[string]any{"limit": 1})
	if out["total_nodes"] != float64(120) || out["total_categories"] != float64(2) || out["matched"] != float64(2) || out["returned"] != float64(1) || out["truncated"] != float64(1) {
		t.Fatalf("bad counts: %+v", out)
	}
	c := out["categories"].([]any)[0].(map[string]any)
	bounds := c["bounds"].(map[string]any)
	if c["category"] != "berries" || c["nodes"] != float64(60) || bounds["min_x"] != float64(0) || bounds["max_x"] != float64(118) || bounds["min_z"] != float64(99) {
		t.Fatalf("bad category: %+v", c)
	}
	out = callQuery(t, cs, ctx, "summarize_resource_coverage", map[string]any{"category": "absent"})
	if out["matched"] != float64(0) || out["matched_nodes"] != float64(0) {
		t.Fatal(out)
	}
}
func TestCoverageUnknownAndEmpty(t *testing.T) {
	for _, nodes := range []any{[]any{}, []any{map[string]any{"id": "unknown", "class_asset_path": "class", "position": []int{-1, 3, 7}}}} {
		cs, ctx := coverageSession(t, map[string]any{"schema": "bellwright.harvestable-baseline.v1", "source_kind": "packaged_game_asset", "nodes": nodes, "unknowns": []string{"availability unknown"}})
		out := callQuery(t, cs, ctx, "summarize_resource_coverage", map[string]any{"category": ""})
		if out["availability"] != "unknown" || out["coordinate_units"] != "engine" {
			t.Fatal(out)
		}
		if out["returned"] == float64(1) && out["categories"].([]any)[0].(map[string]any)["category_known"] != false {
			t.Fatal(out)
		}
	}
}
func TestCoverageInvalidEvidence(t *testing.T) {
	baseline := func() map[string]any {
		return map[string]any{"schema": "bellwright.harvestable-baseline.v1", "source_kind": "packaged_game_asset", "nodes": []any{map[string]any{"id": "a", "class_asset_path": "c", "position": []any{1, 2, 3}}}, "unknowns": []string{"unknown"}}
	}
	for _, field := range []string{"schema", "source_kind", "nodes", "unknowns", "position", "id", "class_asset_path"} {
		t.Run(field, func(t *testing.T) {
			e := baseline()
			if field == "position" || field == "id" || field == "class_asset_path" {
				e["nodes"].([]any)[0].(map[string]any)[field] = nil
			} else {
				e[field] = nil
			}
			cs, ctx := coverageSession(t, e)
			r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "summarize_resource_coverage", Arguments: map[string]any{}})
			if err == nil && !r.IsError {
				t.Fatal("accepted invalid evidence")
			}
		})
	}
}
func TestCoverageInvalidArgumentsAndBackend(t *testing.T) {
	resources, _ := queryFixtures()
	cs, ctx := coverageSession(t, resources)
	for _, limit := range []int{0, 101} {
		r, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "summarize_resource_coverage", Arguments: map[string]any{"limit": limit}})
		if err == nil && !r.IsError {
			t.Fatal("accepted invalid limit")
		}
	}
	h := httptest.NewServer(http.NotFoundHandler())
	b, _ := newBackend(h.URL)
	h.Close()
	if _, _, err := b.resourceCoverage(context.Background(), nil, coverageQuery{}); err == nil {
		t.Fatal("backend down accepted")
	}
}
