package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Exercise the integrated registration and validate actual structured results
// against every advertised output schema, including missing values and stale data.
func TestCapabilityIntegrationSchemas(t *testing.T) {
	raw, err := os.ReadFile("testdata/capabilities.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected mutation %s", r.Method)
		}
		fixture, ok := fixtures[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(fixture)
	}))
	defer h.Close()
	b, err := newBackend(h.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := server(b, b).Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "capability-integration", Version: "1"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()
	args := map[string]map[string]any{
		"read_saved_world": {}, "list_save_snapshots": {}, "read_save_trends": {},
		"read_companion_telemetry": {}, "read_live_map": {}, "read_resource_baseline": {},
		"read_foraging_candidates": {}, "read_gear_plan": {"limit": 1},
		"read_item_history":      {"item_id": "Iron_C", "limit": 1},
		"search_saved_villagers": {"guid": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "limit": 1},
		"locate_saved_inventory": {"item": "Iron_C", "limit": 1},
		"inspect_saved_squads":   {"limit": 1}, "read_saved_village_progress": {"limit": 1},
		"read_connection_health":      {},
		"rank_saved_workers":          {"skill": "craft", "limit": 1},
		"audit_saved_job_priorities":  {"limit": 1},
		"read_saved_housing":          {"limit": 1},
		"compare_saved_stock_targets": {"targets": map[string]any{"Iron_C": 5}, "scope": "stored", "limit": 1},
		"search_observed_items":       {"query": "iron", "limit": 1},
		"summarize_resource_coverage": {"limit": 1},
	}
	list, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != len(args) {
		t.Fatalf("expected %d tools, got %d", len(args), len(list.Tools))
	}
	for _, tool := range list.Tools {
		t.Run(tool.Name, func(t *testing.T) {
			a, ok := args[tool.Name]
			if !ok {
				t.Fatal("unverified registered tool")
			}
			if tool.Annotations == nil || !tool.Annotations.ReadOnlyHint {
				t.Fatal("missing read-only hint")
			}
			out := callQuery(t, cs, ctx, tool.Name, a)
			if tool.Name == "read_gear_plan" {
				if out["snapshot_id"] != float64(7) {
					t.Fatal("lost save provenance")
				}
			}
			if tool.Name == "read_live_map" {
				if out["data"].(map[string]any)["status"].(map[string]any)["stale"] != true {
					t.Fatal("lost stale state")
				}
			}
		})
	}
}
