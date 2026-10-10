package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"strings"
)

type squadQuery struct {
	Name        string `json:"name,omitempty"`
	Limit       *int   `json:"limit,omitempty"`
	MemberLimit *int   `json:"member_limit,omitempty"`
}
type squadMember struct {
	GUID       string            `json:"guid"`
	Resolution string            `json:"resolution"`
	FirstName  *string           `json:"first_name"`
	LastName   *string           `json:"last_name"`
	Recruited  *bool             `json:"is_player_npc"`
	Equipment  map[string]string `json:"equipment"`
}
type squadItem struct {
	GUID             string        `json:"guid"`
	Name             *string       `json:"name"`
	TotalMembers     int           `json:"total_members"`
	ReturnedMembers  int           `json:"returned_members"`
	TruncatedMembers int           `json:"truncated_members"`
	Members          []squadMember `json:"members"`
}
type squadOutput struct {
	Source     string      `json:"source"`
	SnapshotID *int64      `json:"snapshot_id"`
	IngestedAt *string     `json:"ingested_at"`
	State      string      `json:"state"`
	Total      int         `json:"total"`
	Matched    int         `json:"matched"`
	Returned   int         `json:"returned"`
	Truncated  int         `json:"truncated"`
	Groups     []squadItem `json:"groups"`
	Note       string      `json:"note"`
}

func registerSquadTools(s *mcp.Server, b *backend) {
	schema := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"name": map[string]any{"type": "string", "maxLength": 128}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 25}, "member_limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 25}}}
	mcp.AddTool(s, &mcp.Tool{Name: "inspect_saved_squads", InputSchema: schema, Description: "Bounded saved group GUID resolution and equipment. Unmatched or duplicate NPC GUIDs explicit; saved membership does not establish current runtime ownership.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.squads)
}
func (b *backend) squads(ctx context.Context, _ *mcp.CallToolRequest, q squadQuery) (*mcp.CallToolResult, squadOutput, error) {
	o := squadOutput{Source: "companion_save", Groups: []squadItem{}, Note: "Latest ingested save only; ingestion is not game capture time. Group membership and player-NPC flag are saved evidence; current runtime ownership, orders and equipment durability unknown."}
	limit, memberLimit := 25, 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if q.MemberLimit != nil {
		memberLimit = *q.MemberLimit
	}
	if limit < 1 || limit > 100 || memberLimit < 1 || memberLimit > 100 || len(q.Name) > 128 {
		return nil, o, fmt.Errorf("limits must be 1 through 100 and name at most 128 bytes")
	}
	_, raw, e := b.read(ctx, "/api/world", "companion_save", "")
	if e != nil {
		return nil, o, fmt.Errorf("saved world unavailable")
	}
	var w struct {
		SnapshotID *int64  `json:"snapshot_id"`
		IngestedAt *string `json:"ingested_at"`
		Groups     *[]struct {
			GUID    string   `json:"guid"`
			Name    *string  `json:"name"`
			Members []string `json:"members"`
		} `json:"groups"`
		NPCs []struct {
			GUID      *string           `json:"guid"`
			FirstName *string           `json:"first_name"`
			LastName  *string           `json:"last_name"`
			Recruited *bool             `json:"is_player_npc"`
			Equipment map[string]string `json:"equipment"`
		} `json:"npcs"`
	}
	if decodeEvidence(raw.Data, &w) != nil {
		return nil, o, fmt.Errorf("invalid saved world evidence")
	}
	o.SnapshotID = w.SnapshotID
	o.IngestedAt = w.IngestedAt
	if w.Groups == nil {
		o.State = "missing"
		return nil, o, nil
	}
	o.State = "available"
	o.Total = len(*w.Groups)
	if o.Total == 0 {
		o.State = "empty"
	}
	index := map[string][]squadMember{}
	for _, n := range w.NPCs {
		if n.GUID != nil {
			index[*n.GUID] = append(index[*n.GUID], squadMember{GUID: *n.GUID, Resolution: "matched", FirstName: n.FirstName, LastName: n.LastName, Recruited: n.Recruited, Equipment: n.Equipment})
		}
	}
	for _, g := range *w.Groups {
		if q.Name != "" && (g.Name == nil || !strings.Contains(strings.ToLower(*g.Name), strings.ToLower(q.Name))) {
			continue
		}
		o.Matched++
		if len(o.Groups) >= limit {
			continue
		}
		item := squadItem{GUID: g.GUID, Name: g.Name, TotalMembers: len(g.Members), Members: []squadMember{}}
		for _, id := range g.Members {
			if len(item.Members) >= memberLimit {
				break
			}
			m := squadMember{GUID: id, Resolution: "unmatched"}
			matches := index[id]
			if len(matches) == 1 {
				m = matches[0]
			} else if len(matches) > 1 {
				m.Resolution = "ambiguous"
			}
			item.Members = append(item.Members, m)
		}
		item.ReturnedMembers = len(item.Members)
		item.TruncatedMembers = item.TotalMembers - item.ReturnedMembers
		o.Groups = append(o.Groups, item)
	}
	o.Returned = len(o.Groups)
	o.Truncated = o.Matched - o.Returned
	return nil, o, nil
}
