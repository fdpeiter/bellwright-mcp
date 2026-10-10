package main

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

func TestJobPrioritiesMCP(t *testing.T) {
	npcs := []any{}
	for _, jobs := range []any{map[string]any{"farming": 1, "cooking": 1, "hunting": 2}, map[string]any{}, nil, map[string]any{"farming": 10}, map[string]any{"farming": "bad"}, map[string]any{"farming": 0}} {
		npcs = append(npcs, map[string]any{"is_player_npc": true, "job_priorities": jobs})
	}
	npcs = append(npcs, map[string]any{"is_player_npc": false, "job_priorities": map[string]any{"a": 1, "b": 1}})
	cs, ctx := workerFitSession(t, map[string]any{"npcs": npcs}, registerJobPriorityTools)
	o := callQuery(t, cs, ctx, "audit_saved_job_priorities", map[string]any{"limit": 1})
	if o["total"] != float64(6) || o["observed_ties"] != float64(1) || o["observed_empty_priority_maps"] != float64(1) || o["unknown_priority_maps"] != float64(1) || o["malformed_priority_maps"] != float64(2) || o["truncated"] != float64(5) {
		t.Fatal(o)
	}
	i := o["items"].([]any)[0].(map[string]any)
	if i["strongest_observed_priority"] != float64(1) || len(i["strongest_observed_jobs"].([]any)) != 2 {
		t.Fatal(i)
	}
	for _, limit := range []int{0, 101} {
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "audit_saved_job_priorities", Arguments: map[string]any{"limit": limit}})
		if e == nil && !r.IsError {
			t.Fatal(limit)
		}
	}
	o = callQuery(t, cs, ctx, "audit_saved_job_priorities", map[string]any{"limit": 100})
	if o["returned"] != float64(6) {
		t.Fatal(o)
	}
}
func TestJobPrioritiesMissingMalformed(t *testing.T) {
	for _, world := range []any{map[string]any{}, map[string]any{"npcs": false}} {
		cs, ctx := workerFitSession(t, world, registerJobPriorityTools)
		r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "audit_saved_job_priorities", Arguments: map[string]any{}})
		if e == nil && !r.IsError {
			t.Fatal(world)
		}
	}
	cs, ctx := workerFitSession(t, map[string]any{"npcs": []any{map[string]any{"is_player_npc": true}}}, registerJobPriorityTools)
	o := callQuery(t, cs, ctx, "audit_saved_job_priorities", map[string]any{})
	if o["unknown_priority_maps"] != float64(1) || o["limit"] != float64(25) {
		t.Fatal(o)
	}
}
