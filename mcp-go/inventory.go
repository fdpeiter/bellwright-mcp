package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
)

type inventoryQuery struct {
	Item        string `json:"item,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	Faction     string `json:"faction,omitempty"`
	Limit       *int   `json:"limit,omitempty"`
}
type inventoryOutput struct {
	Source      string             `json:"source"`
	SnapshotID  *int               `json:"snapshot_id"`
	IngestedAt  *string            `json:"ingested_at"`
	RetrievedAt string             `json:"retrieved_at"`
	Total       int                `json:"total"`
	Matched     int                `json:"matched"`
	Returned    int                `json:"returned"`
	Truncated   int                `json:"truncated"`
	Limit       int                `json:"limit"`
	Containers  []map[string]any   `json:"containers"`
	Storage     map[string]float64 `json:"storage"`
	NPCCarried  map[string]float64 `json:"npc_carried"`
	Equipped    map[string]int     `json:"equipped"`
	Note        string             `json:"note"`
}

func registerInventoryTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "locate_saved_inventory", Description: "Locate exact saved item class and/or container ID. Optional exact faction filter. Returns bounded containers with saved quantities and positions; aggregate storage, NPC-carried and equipped scopes stay separate. Unknown ownership and position remain unknown. Limit defaults 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.inventory)
}
func inventoryQuantities(m map[string]float64, item string) (map[string]float64, error) {
	if m == nil {
		return nil, nil
	}
	out := map[string]float64{}
	for k, v := range m {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return nil, fmt.Errorf("invalid item quantity")
		}
		if item == "" || k == item {
			out[k] = v
		}
	}
	return out, nil
}
func (b *backend) inventory(ctx context.Context, _ *mcp.CallToolRequest, q inventoryQuery) (*mcp.CallToolResult, inventoryOutput, error) {
	out := inventoryOutput{}
	limit := 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > 100 {
		return nil, out, fmt.Errorf("limit must be 1 through 100")
	}
	for _, v := range []string{q.Item, q.ContainerID, q.Faction} {
		if len(v) > 256 {
			return nil, out, fmt.Errorf("filter exceeds 256 bytes")
		}
	}
	_, raw, err := b.read(ctx, "/api/world", "companion_save", "")
	if err != nil {
		return nil, out, err
	}
	var world struct {
		SnapshotID *int    `json:"snapshot_id"`
		IngestedAt *string `json:"ingested_at"`
		Storage    *struct {
			Totals     map[string]float64 `json:"totals"`
			Carried    map[string]float64 `json:"npc_carried"`
			Containers []map[string]any   `json:"containers"`
		} `json:"storage"`
		NPCs []map[string]any `json:"npcs"`
	}
	if err = decodeEvidence(raw.Data, &world); err != nil {
		return nil, out, err
	}
	if world.Storage == nil || world.Storage.Containers == nil || len(world.Storage.Containers) > 100000 {
		return nil, out, fmt.Errorf("missing or invalid storage evidence")
	}
	out = inventoryOutput{Source: raw.Source, SnapshotID: world.SnapshotID, IngestedAt: world.IngestedAt, RetrievedAt: raw.RetrievedAt, Total: len(world.Storage.Containers), Limit: limit, Containers: []map[string]any{}, Note: "Latest ingested save only; saved positions are not live locations. Container/faction filters affect container results only; aggregate scopes cover backend evidence and item filter only. Storage totals and NPC-carried quantities are independent source aggregates. Equipped counts are recruited NPC slots, not storage stock; missing NPC evidence remains unknown. Ownership derives only from observed faction; no faction means unknown."}
	out.Storage, err = inventoryQuantities(world.Storage.Totals, q.Item)
	if err != nil {
		return nil, inventoryOutput{}, err
	}
	out.NPCCarried, err = inventoryQuantities(world.Storage.Carried, q.Item)
	if err != nil {
		return nil, inventoryOutput{}, err
	}
	if world.NPCs != nil {
		out.Equipped = map[string]int{}
		for _, n := range world.NPCs {
			if n["is_player_npc"] != true {
				continue
			}
			equipment, ok := n["equipment"].(map[string]any)
			if !ok {
				out.Equipped = nil
				break
			}
			for _, v := range equipment {
				item, ok := v.(string)
				if !ok {
					return nil, inventoryOutput{}, fmt.Errorf("invalid equipped evidence")
				}
				if item != "" && (q.Item == "" || q.Item == item) {
					out.Equipped[item]++
				}
			}
		}
	}
	for _, c := range world.Storage.Containers {
		id, ok := c["id"].(string)
		if !ok || id == "" {
			return nil, inventoryOutput{}, fmt.Errorf("missing container ID")
		}
		items, ok := c["items"].(map[string]any)
		if !ok {
			return nil, inventoryOutput{}, fmt.Errorf("missing container items")
		}
		for _, v := range items {
			n, ok := v.(float64)
			if !ok || n < 0 {
				return nil, inventoryOutput{}, fmt.Errorf("invalid container quantity")
			}
		}
		if pos, exists := c["position"]; exists && pos != nil {
			p, ok := pos.([]any)
			if !ok || len(p) != 3 {
				return nil, inventoryOutput{}, fmt.Errorf("invalid saved position")
			}
			for _, v := range p {
				n, ok := v.(float64)
				if !ok || !validCoordinate(n) {
					return nil, inventoryOutput{}, fmt.Errorf("invalid saved coordinate")
				}
			}
		}
		if q.ContainerID != "" && id != q.ContainerID {
			continue
		}
		if q.Faction != "" && c["faction"] != q.Faction {
			continue
		}
		if q.Item != "" {
			v, ok := items[q.Item].(float64)
			if !ok || v <= 0 {
				continue
			}
		}
		out.Matched++
		if len(out.Containers) < limit {
			out.Containers = append(out.Containers, c)
		}
	}
	out.Returned = len(out.Containers)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
