package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

func TestVillageEvidenceAndBounds(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"snapshot_id":10,"ingested_at":"ingestion","villages":[{"name":"Haerndean","trust":0,"trust_level":0,"prosperity":0,"liberated":false},{"name":"Haerndean extension"},{"name":"Other","trust":20,"liberated":true}]}`})
	one := 1
	_, o, e := b.villageProgress(context.Background(), nil, villageQuery{Name: "HAERN", Limit: &one})
	if e != nil || o.Total != 3 || o.Matched != 2 || o.Returned != 1 || o.Truncated != 1 || o.Villages[0].Trust == nil || *o.Villages[0].Liberated {
		t.Fatalf("%+v %v", o, e)
	}
	_, o, e = b.villageProgress(context.Background(), nil, villageQuery{Name: "extension"})
	if e != nil || o.Villages[0].Trust != nil || o.Villages[0].Liberated != nil {
		t.Fatalf("unknowns: %+v %v", o, e)
	}
}
func TestVillageMissingEmptyNoMatches(t *testing.T) {
	for _, tc := range []struct{ raw, state string }{{`{}`, "missing"}, {`{"villages":[]}`, "empty"}, {`{"villages":[{"name":"Other"}]}`, "available"}} {
		b := healthFixture(t, map[string]string{"/api/world": tc.raw})
		_, o, e := b.villageProgress(context.Background(), nil, villageQuery{Name: "absent"})
		if e != nil || o.State != tc.state || o.Matched != 0 {
			t.Fatalf("%+v %v", o, e)
		}
	}
}
func TestVillageMCP(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"villages":[]}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "village-test", Version: "1"}, nil)
	registerVillageTools(s, b)
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
	list, e := cs.ListTools(ctx, nil)
	if e != nil || list.Tools[0].OutputSchema == nil {
		t.Fatal("missing output schema")
	}
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_saved_village_progress", Arguments: map[string]any{"limit": 1, "name": "test"}})
	if e != nil || r.IsError {
		t.Fatalf("%v %v", r, e)
	}
	r, e = cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_saved_village_progress", Arguments: map[string]any{"limit": 0}})
	if e == nil && !r.IsError {
		t.Fatal("schema accepted zero bound")
	}
}
