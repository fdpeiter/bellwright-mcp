package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVillagers(t *testing.T) {
	fetched := 0
	fixture := `{"snapshot_id":7,"ingested_at":"saved","npcs":[{"is_player_npc":true,"guid":"a","first_name":"Ada","morale":null,"skills":{"craft":{"level":4}},"injuries":[],"housed":null},{"is_player_npc":true,"guid":"b","morale":42,"injuries":["hurt"]},{"is_player_npc":false}]}`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetched++
		if r.URL.Path != "/api/world" {
			t.Error(r.URL.Path)
		}
		w.Write([]byte(fixture))
	}))
	defer s.Close()
	b, _ := newBackend(s.URL)
	limit := 1
	_, out, e := b.villagers(context.Background(), nil, villagerQuery{Limit: &limit})
	if e != nil || out.Total != 2 || out.Truncated != 1 || out.SnapshotID == nil || *out.SnapshotID != 7 {
		t.Fatalf("%+v %v", out, e)
	}
	min := 0.0
	_, out, e = b.villagers(context.Background(), nil, villagerQuery{MinMorale: &min})
	if e != nil || out.Matched != 1 {
		t.Fatalf("%+v %v", out, e)
	}
	_, out, e = b.villagers(context.Background(), nil, villagerQuery{Name: "ADA", Skill: "craft"})
	if e != nil || out.Matched != 1 || out.Items[0]["morale"] != nil {
		t.Fatalf("%+v %v", out, e)
	}
	before := fetched
	limit = 101
	_, _, e = b.villagers(context.Background(), nil, villagerQuery{Limit: &limit})
	if e == nil || fetched != before {
		t.Fatal("invalid fetched")
	}
	for _, f := range []string{`{}`, `{"npcs":null}`, `{"npcs":42}`} {
		fixture = f
		_, _, e = b.villagers(context.Background(), nil, villagerQuery{})
		if e == nil {
			t.Fatalf("accepted %s", f)
		}
	}
	fixture = `{"npcs":[]}`
	_, out, e = b.villagers(context.Background(), nil, villagerQuery{})
	if e != nil || out.Items == nil || out.SnapshotID != nil {
		t.Fatalf("%+v %v", out, e)
	}
}
func TestVillagerSchemas(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	registerVillagerTools(s, &backend{})
	ctx := context.Background()
	a, b := mcp.NewInMemoryTransports()
	session, e := s.Connect(ctx, a, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer session.Close()
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	client, e := c.Connect(ctx, b, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer client.Close()
	list, e := client.ListTools(ctx, nil)
	if e != nil || len(list.Tools) != 1 {
		t.Fatal(e)
	}
	tool := list.Tools[0]
	if tool.InputSchema == nil || tool.OutputSchema == nil || !tool.Annotations.ReadOnlyHint {
		t.Fatal("missing schemas")
	}
}
