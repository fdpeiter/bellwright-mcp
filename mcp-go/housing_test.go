package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"testing"
)

func TestHousingEvidence(t *testing.T) {
	one := 1
	b := healthFixture(t, map[string]string{"/api/world": `{"snapshot_id":2,"ingested_at":"old","housing":{"quarters":2,"houses":1,"byType":[{"cls":"Tent","label":"Tent","cap":2,"count":1,"beds":2}]},"npcs":[{"guid":"a","is_player_npc":true,"housed":false},{"guid":"b","is_player_npc":true},{"is_player_npc":true,"housed":true},{"housed":false},{"is_player_npc":false,"housed":false}]}`})
	_, o, e := b.savedHousing(context.Background(), nil, housingQuery{Limit: &one})
	if e != nil || *o.CapacityDeficit != 1 || o.ObservedNotHoused != 1 || o.UnknownHousing != 1 || o.UnknownOwnership != 1 || o.IndividualTruncated != 2 || *o.IngestedAt != "old" {
		t.Fatalf("%+v %v", o, e)
	}
}
func TestHousingMissingZeroInvalid(t *testing.T) {
	for _, tc := range []struct {
		raw, state string
		bad        bool
	}{{`{}`, "missing", false}, {`{"housing":null}`, "missing", false}, {`{"housing":{}}`, "unknown", false}, {`{"housing":{"quarters":0,"houses":0,"byType":[]},"npcs":[]}`, "available", false}, {`{"housing":{"quarters":null,"houses":0}}`, "unknown", false}, {`{"housing":{"quarters":0.5,"houses":0}}`, "", true}, {`{"housing":{"quarters":-1,"houses":0}}`, "", true}, {`{"housing":{"quarters":2,"houses":0,"byType":[]}}`, "", true}, {`{"housing":{"quarters":2,"houses":1,"byType":[{"cls":"x","cap":2,"count":1,"beds":1}]}}`, "", true}, {`{"npcs":[{"is_player_npc":true,"housed":"false"}]}`, "", true}} {
		b := healthFixture(t, map[string]string{"/api/world": tc.raw})
		_, o, e := b.savedHousing(context.Background(), nil, housingQuery{})
		if (e != nil) != tc.bad || !tc.bad && o.State != tc.state {
			t.Fatalf("%s: %+v %v", tc.raw, o, e)
		}
		if tc.state == "missing" && (o.Quarters != nil || o.CapacityDeficit != nil) {
			t.Fatal("missing became zero")
		}
	}
}
func TestHousingMCP(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world": `{"housing":{"quarters":0,"houses":0},"npcs":[{"is_player_npc":true,"housed":false}]}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "housing-test", Version: "1"}, nil)
	registerHousingTools(s, b)
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
	o := callQuery(t, cs, ctx, "read_saved_housing", map[string]any{"limit": 1})
	if o["capacity_deficit"] != float64(1) {
		t.Fatal(o)
	}
	b2 := 0
	_, _, e = b.savedHousing(ctx, nil, housingQuery{Limit: &b2})
	if e == nil {
		t.Fatal("accepted invalid limit")
	}
}
