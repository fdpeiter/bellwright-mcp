// Isolated MCP extension for food observations while the shared bridge is in flight.
package main

import (
	"context"
	"flag"
	"github.com/claygorman/bellwright-companion/mcp-go/internal/foodcontrol"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"log"
	"time"
)

type input struct{}

func newServer(path string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "bellwright-steward-food", Version: "0.1.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "read_live_food_state", Description: "Read actual live food-building orders, storage rules and observed settlement membership. Missing targets and unknown ownership remain unknown. No saved-data fallback or game writes.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, foodcontrol.Result, error) {
		v, e := foodcontrol.Read(path, time.Now().UTC())
		return nil, v, e
	})
	mcp.AddTool(s, &mcp.Tool{Name: "read_live_storage_state", Description: "Read bounded live stockpile and chest inventories, settlement faction references, native capacity checks and access flags. Nearby containers can belong to other factions or have unknown ownership. No saved-data fallback or transfers.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, foodcontrol.Result, error) {
		v, e := foodcontrol.Read(path, time.Now().UTC())
		v.Note = "Live storage observations are in data.storage. Faction references and access flags do not alone prove transfer authorization; capacity checks sample existing items. Transfers remain disabled."
		if v.Available && (v.Data == nil || v.Data.Storage == nil) {
			v.Available = false
			v.Status = "storage_unavailable"
		}
		return nil, v, e
	})
	return s
}
func main() {
	path := flag.String("log", "", "Private UE4SS runtime log path")
	flag.Parse()
	if *path == "" {
		log.Fatal("-log is required; no save fallback")
	}
	if e := newServer(*path).Run(context.Background(), &mcp.StdioTransport{}); e != nil {
		log.Fatal(e)
	}
}
