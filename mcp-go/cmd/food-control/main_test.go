package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestToolSchema(t *testing.T) {
	ctx := context.Background()
	a, b := mcp.NewInMemoryTransports()
	path := filepath.Join(t.TempDir(), "runtime.log")
	line := fmt.Sprintf(`BWFOOD {"source":"ue4ss_live","session_id":"schema-test","captured_at":%q,"available":true,"status":"observed","buildings":[{"orders":[],"orders_status":{"available":true}}],"settlements":[{"residents":[]}],"containers":[{"slots":[{"item":"Mushroom_C","quantity":2}]}]}`, time.Now().UTC().Format(time.RFC3339))
	if err := os.WriteFile(path, []byte(line), 0600); err != nil {
		t.Fatal(err)
	}
	s, err := newServer(path).Connect(ctx, a, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, b, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	list, err := c.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Tools) != 1 || list.Tools[0].Name != "read_live_food_state" || list.Tools[0].OutputSchema == nil || !list.Tools[0].Annotations.ReadOnlyHint {
		t.Fatalf("unexpected tool schema: %#v", list)
	}
	result, err := c.CallTool(ctx, &mcp.CallToolParams{Name: "read_live_food_state", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatalf("populated schema call failed: %#v %v", result, err)
	}
}
