package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

func TestSquadResolutionAndBounds(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"snapshot_id":9,"ingested_at":"ingestion","groups":[{"guid":"g","name":"Guard","members":["known","unknown","duplicate","overflow"]},{"guid":"g2","name":"Guard Two","members":[]}],"npcs":[{"guid":"known","first_name":"A","is_player_npc":true,"equipment":{"head":"Hat"}},{"guid":"duplicate"},{"guid":"duplicate"}]}`})
	one, three := 1, 3
	_, o, e := b.squads(context.Background(), nil, squadQuery{Name: "guard", Limit: &one, MemberLimit: &three})
	if e != nil || o.Matched != 2 || o.Truncated != 1 || o.Groups[0].TruncatedMembers != 1 || o.Groups[0].Members[0].Equipment["head"] != "Hat" || o.Groups[0].Members[1].Resolution != "unmatched" || o.Groups[0].Members[2].Resolution != "ambiguous" {
		t.Fatalf("%+v %v", o, e)
	}
	zero := 0
	if _, _, e = b.squads(context.Background(), nil, squadQuery{MemberLimit: &zero}); e == nil {
		t.Fatal("accepted zero member limit")
	}
}
func TestSquadMissingVsEmpty(t *testing.T) {
	for _, tc := range []struct{ raw, state string }{{`{"npcs":[]}`, "missing"}, {`{"groups":[],"npcs":[]}`, "empty"}} {
		b := healthFixture(t, map[string]string{"/api/world": tc.raw})
		_, o, e := b.squads(context.Background(), nil, squadQuery{})
		if e != nil || o.State != tc.state {
			t.Fatalf("%+v %v", o, e)
		}
	}
}
func TestSquadMCP(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"groups":[],"npcs":[]}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "squad-test", Version: "1"}, nil)
	registerSquadTools(s, b)
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
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "inspect_saved_squads", Arguments: map[string]any{"limit": 1, "member_limit": 2}})
	if e != nil || r.IsError {
		t.Fatalf("%v %v", r, e)
	}
	r, e = cs.CallTool(ctx, &mcp.CallToolParams{Name: "inspect_saved_squads", Arguments: map[string]any{"limit": 101}})
	if e == nil && !r.IsError {
		t.Fatal("schema accepted invalid bound")
	}
}
