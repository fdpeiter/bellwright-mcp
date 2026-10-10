package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type gearQuery struct {
	Section string `json:"section,omitempty"`
	Limit   *int   `json:"limit,omitempty"`
}
type gearCraft struct {
	Item     string `json:"item"`
	Need     int    `json:"need"`
	Equipped int    `json:"equipped"`
	Stored   int    `json:"stored"`
	Reserve  int    `json:"reserve"`
	Target   int    `json:"target"`
	Have     int    `json:"have"`
	Deficit  int    `json:"deficit"`
}
type gearReequip struct {
	Name string  `json:"name"`
	Slot string  `json:"slot"`
	Worn *string `json:"worn"`
	Pick string  `json:"pick"`
}
type gearPreset struct {
	Key         string            `json:"key"`
	Name        string            `json:"name"`
	MemberCount int               `json:"member_count"`
	Picks       map[string]string `json:"picks"`
}
type gearSection[T any] struct {
	Matched   int `json:"matched"`
	Returned  int `json:"returned"`
	Truncated int `json:"truncated"`
	Items     []T `json:"items"`
}
type gearOutput struct {
	Source      string                   `json:"source"`
	RetrievedAt string                   `json:"retrieved_at"`
	SnapshotID  *int                     `json:"snapshot_id"`
	IngestedAt  *string                  `json:"ingested_at"`
	Limit       int                      `json:"limit"`
	HasReserves bool                     `json:"has_reserves"`
	Craft       gearSection[gearCraft]   `json:"craft"`
	Reequip     gearSection[gearReequip] `json:"reequip"`
	Presets     gearSection[gearPreset]  `json:"presets"`
	Note        string                   `json:"note"`
}

func gearBound[T any](items []T, limit int, enabled bool) gearSection[T] {
	o := gearSection[T]{Items: []T{}}
	if !enabled {
		return o
	}
	o.Matched = len(items)
	n := len(items)
	if n > limit {
		n = limit
	}
	o.Items = append(o.Items, items[:n]...)
	o.Returned = n
	o.Truncated = o.Matched - n
	return o
}
func registerGearTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "read_gear_plan", Description: "Bounded saved gear planning. Tier-unlocked is an inventory heuristic, not research unlock evidence.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}, InputSchema: map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"section": map[string]any{"type": "string", "enum": []string{"all", "craft", "reequip", "presets"}, "default": "all"}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 25}}}}, b.gearPlan)
}
func (b *backend) gearPlan(ctx context.Context, _ *mcp.CallToolRequest, q gearQuery) (*mcp.CallToolResult, gearOutput, error) {
	var o gearOutput
	limit, e := (spatialQuery{Limit: q.Limit}).validate()
	if e != nil {
		return nil, o, e
	}
	if q.Section == "" {
		q.Section = "all"
	}
	if q.Section != "all" && q.Section != "craft" && q.Section != "reequip" && q.Section != "presets" {
		return nil, o, fmt.Errorf("invalid section")
	}
	_, raw, e := b.read(ctx, "/api/gear", "companion_save", "")
	if e != nil {
		return nil, o, e
	}
	var d struct {
		SnapshotID *int          `json:"snapshot_id"`
		IngestedAt *string       `json:"ingested_at"`
		Craft      []gearCraft   `json:"craft"`
		Reequip    []gearReequip `json:"reequip"`
		Presets    []struct {
			Key     string            `json:"key"`
			Name    string            `json:"name"`
			Members []struct{}        `json:"members"`
			Picks   map[string]string `json:"picks"`
		} `json:"presets"`
		HasReserves bool `json:"has_reserves"`
	}
	if e = decodeEvidence(raw.Data, &d); e != nil {
		return nil, o, e
	}
	if d.Craft == nil || d.Reequip == nil || d.Presets == nil {
		return nil, o, fmt.Errorf("missing gear evidence arrays")
	}
	presets := []gearPreset{}
	for _, p := range d.Presets {
		if p.Key == "" || p.Members == nil || p.Picks == nil || len(p.Picks) > 8 {
			return nil, o, fmt.Errorf("invalid preset evidence")
		}
		presets = append(presets, gearPreset{p.Key, p.Name, len(p.Members), p.Picks})
	}
	for _, c := range d.Craft {
		if c.Item == "" || c.Deficit <= 0 || c.Stored < 0 || c.Need < 0 || c.Equipped < 0 || c.Reserve < 0 || c.Target < 0 || c.Have < 0 {
			return nil, o, fmt.Errorf("invalid craft evidence")
		}
	}
	for _, r := range d.Reequip {
		if r.Slot == "" || r.Pick == "" {
			return nil, o, fmt.Errorf("invalid reequip evidence")
		}
	}
	o = gearOutput{Source: raw.Source, RetrievedAt: raw.RetrievedAt, SnapshotID: d.SnapshotID, IngestedAt: d.IngestedAt, Limit: limit, HasReserves: d.HasReserves, Craft: gearBound(d.Craft, limit, q.Section == "all" || q.Section == "craft"), Reequip: gearBound(d.Reequip, limit, q.Section == "all" || q.Section == "reequip"), Presets: gearBound(presets, limit, q.Section == "all" || q.Section == "presets"), Note: "Latest ingested save, not live state. Retrieval time is not save time; null snapshot metadata means unknown. Each selected section is bounded independently in backend order. Tier-unlocked means stored > 0 or equipped by at least two villagers; when none qualify the ideal preset item is used. This does not establish actual research/crafting unlocks. Presets omit ranked chains and member identities. Stored stock may be shared across re-equipping candidates."}
	return nil, o, nil
}
