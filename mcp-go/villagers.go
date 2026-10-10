package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
	"strings"
)

type villagerQuery struct {
	Name      string   `json:"name,omitempty"`
	GUID      string   `json:"guid,omitempty"`
	Skill     string   `json:"skill,omitempty"`
	MinLevel  *float64 `json:"min_level,omitempty"`
	Injured   *bool    `json:"injured,omitempty"`
	Housed    *bool    `json:"housed,omitempty"`
	MinMorale *float64 `json:"min_morale,omitempty"`
	MaxMorale *float64 `json:"max_morale,omitempty"`
	Job       string   `json:"job,omitempty"`
	Priority  *float64 `json:"priority,omitempty"`
	Limit     *int     `json:"limit,omitempty"`
}
type villagerOutput struct {
	Source      string           `json:"source"`
	SnapshotID  *int             `json:"snapshot_id"`
	IngestedAt  *string          `json:"ingested_at"`
	RetrievedAt string           `json:"retrieved_at"`
	Total       int              `json:"total"`
	Matched     int              `json:"matched"`
	Returned    int              `json:"returned"`
	Truncated   int              `json:"truncated"`
	Limit       int              `json:"limit"`
	Items       []map[string]any `json:"items"`
	Note        string           `json:"note"`
}

func registerVillagerTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "search_saved_villagers", Description: "Search recruited NPCs explicitly marked is_player_npc in the latest ingested save. Missing fields remain unknown and do not satisfy filters. Name is case-insensitive substring; GUID, skill and job keys are exact. Limit defaults to 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.villagers)
}
func (b *backend) villagers(ctx context.Context, _ *mcp.CallToolRequest, q villagerQuery) (*mcp.CallToolResult, villagerOutput, error) {
	out := villagerOutput{}
	limit := 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > 100 {
		return nil, out, fmt.Errorf("limit must be 1 through 100")
	}
	for _, v := range []string{q.Name, q.GUID, q.Skill, q.Job} {
		if len(v) > 128 {
			return nil, out, fmt.Errorf("filter exceeds 128 bytes")
		}
	}
	if q.MinLevel != nil && q.Skill == "" || q.Priority != nil && q.Job == "" {
		return nil, out, fmt.Errorf("min_level requires skill; priority requires job")
	}
	for _, v := range []*float64{q.MinLevel, q.MinMorale, q.MaxMorale, q.Priority} {
		if v != nil && (math.IsNaN(*v) || math.IsInf(*v, 0) || math.Abs(*v) > 1e6) {
			return nil, out, fmt.Errorf("invalid numeric filter")
		}
	}
	if q.MinLevel != nil && *q.MinLevel < 0 {
		return nil, out, fmt.Errorf("min_level must be nonnegative")
	}
	if q.MinMorale != nil && q.MaxMorale != nil && *q.MinMorale > *q.MaxMorale {
		return nil, out, fmt.Errorf("morale minimum exceeds maximum")
	}
	_, raw, err := b.read(ctx, "/api/world", "companion_save", "")
	if err != nil {
		return nil, out, err
	}
	var world struct {
		SnapshotID *int             `json:"snapshot_id"`
		IngestedAt *string          `json:"ingested_at"`
		NPCs       []map[string]any `json:"npcs"`
	}
	if err = decodeEvidence(raw.Data, &world); err != nil {
		return nil, out, err
	}
	if world.NPCs == nil || len(world.NPCs) > 100000 {
		return nil, out, fmt.Errorf("missing or invalid saved NPC evidence")
	}
	out = villagerOutput{Source: raw.Source, SnapshotID: world.SnapshotID, IngestedAt: world.IngestedAt, RetrievedAt: raw.RetrievedAt, Limit: limit, Items: []map[string]any{}, Note: "Latest ingested save only; retrieval time is not save time. Total counts explicit recruited NPC records, not runtime actors. Unknown values never satisfy filters; backend order preserved."}
	for _, n := range world.NPCs {
		recruited, ok := n["is_player_npc"].(bool)
		if !ok || !recruited {
			continue
		}
		out.Total++
		guid, _ := n["guid"].(string)
		if q.GUID != "" && guid != q.GUID {
			continue
		}
		first, _ := n["first_name"].(string)
		last, _ := n["last_name"].(string)
		if q.Name != "" && !strings.Contains(strings.ToLower(first+" "+last), strings.ToLower(q.Name)) {
			continue
		}
		if q.Housed != nil {
			v, ok := n["housed"].(bool)
			if !ok || v != *q.Housed {
				continue
			}
		}
		if q.Injured != nil {
			v, ok := n["injuries"].([]any)
			if !ok || (len(v) > 0) != *q.Injured {
				continue
			}
		}
		if q.MinMorale != nil || q.MaxMorale != nil {
			v, ok := n["morale"].(float64)
			if !ok || q.MinMorale != nil && v < *q.MinMorale || q.MaxMorale != nil && v > *q.MaxMorale {
				continue
			}
		}
		if q.Skill != "" {
			skills, ok := n["skills"].(map[string]any)
			if !ok {
				continue
			}
			s, ok := skills[q.Skill].(map[string]any)
			if !ok {
				continue
			}
			v, ok := s["level"].(float64)
			if !ok || q.MinLevel != nil && v < *q.MinLevel {
				continue
			}
		}
		if q.Job != "" {
			jobs, ok := n["job_priorities"].(map[string]any)
			if !ok {
				continue
			}
			v, ok := jobs[q.Job].(float64)
			if !ok || q.Priority != nil && v != *q.Priority {
				continue
			}
		}
		out.Matched++
		if len(out.Items) < limit {
			out.Items = append(out.Items, n)
		}
	}
	out.Returned = len(out.Items)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
