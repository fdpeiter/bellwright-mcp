package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInventory(t *testing.T) {
	calls := 0
	fixture := `{"snapshot_id":1,"storage":{"totals":{"iron":3},"npc_carried":{"iron":2},"containers":[{"id":"a","faction":null,"position":null,"items":{"iron":3}},{"id":"b","items":{}}]},"npcs":[{"is_player_npc":true,"equipment":{"hand":"iron"}}]}`
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/api/world" {
			t.Error(r.URL.Path)
		}
		w.Write([]byte(fixture))
	}))
	defer s.Close()
	b, _ := newBackend(s.URL)
	_, out, e := b.inventory(context.Background(), nil, inventoryQuery{Item: "iron"})
	if e != nil || out.Total != 2 || out.Matched != 1 || out.Storage["iron"] != 3 || out.NPCCarried["iron"] != 2 || out.Equipped["iron"] != 1 || out.Containers[0]["faction"] != nil {
		t.Fatalf("%+v %v", out, e)
	}
	_, out, e = b.inventory(context.Background(), nil, inventoryQuery{Faction: "Player"})
	if e != nil || out.Matched != 0 || out.Storage["iron"] != 3 {
		t.Fatalf("%+v %v", out, e)
	}
	limit := 0
	before := calls
	_, _, e = b.inventory(context.Background(), nil, inventoryQuery{Limit: &limit})
	if e == nil || calls != before {
		t.Fatal("invalid input fetched")
	}
	for _, f := range []string{`{}`, `{"storage":{"containers":null}}`, `{"storage":{"containers":[{"id":"a","items":{"iron":null}}]}}`, `{"storage":{"containers":[{"id":"a","items":{},"position":[1]}]}}`} {
		fixture = f
		_, _, e = b.inventory(context.Background(), nil, inventoryQuery{})
		if e == nil {
			t.Fatal(f)
		}
	}
	fixture = `{"storage":{"containers":[]}}`
	_, out, e = b.inventory(context.Background(), nil, inventoryQuery{})
	if e != nil || out.Containers == nil || out.Storage != nil || out.Equipped != nil {
		t.Fatalf("%+v %v", out, e)
	}
}
func TestInventorySchemas(t *testing.T) {
	s := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	registerInventoryTools(s, &backend{})
	ctx := context.Background()
	a, b := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, a, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer ss.Close()
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	cs, e := c.Connect(ctx, b, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer cs.Close()
	list, e := cs.ListTools(ctx, nil)
	if e != nil || len(list.Tools) != 1 {
		t.Fatal(e)
	}
	tool := list.Tools[0]
	if tool.InputSchema == nil || tool.OutputSchema == nil || !tool.Annotations.ReadOnlyHint {
		t.Fatal("missing schemas")
	}
}
