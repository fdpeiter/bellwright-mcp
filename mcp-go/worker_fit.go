package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"sort"
)

type workerFitQuery struct {
	Skill string `json:"skill"`
	Limit *int   `json:"limit,omitempty"`
}
type workerFitItem struct {
	Rank         int            `json:"rank"`
	Tied         int            `json:"tied_workers"`
	Score        float64        `json:"score"`
	Observations map[string]any `json:"observations"`
}
type workerFitOutput struct {
	Source      string          `json:"source"`
	SnapshotID  *int            `json:"snapshot_id"`
	IngestedAt  *string         `json:"ingested_at"`
	RetrievedAt string          `json:"retrieved_at"`
	Skill       string          `json:"skill"`
	Total       int             `json:"total"`
	Unknown     int             `json:"excluded_unknown_skill"`
	Malformed   int             `json:"excluded_malformed_skill"`
	Matched     int             `json:"matched"`
	Returned    int             `json:"returned"`
	Truncated   int             `json:"truncated"`
	Limit       int             `json:"limit"`
	Items       []workerFitItem `json:"items"`
	Note        string          `json:"note"`
}

func registerWorkerFitTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "rank_saved_workers", Description: "Advisory recruited-worker ranking by one exact observed skill level. No trait modifiers or unverified job mappings; ties retain save order. Default limit 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.workerFit)
}
func (b *backend) workerFit(ctx context.Context, _ *mcp.CallToolRequest, q workerFitQuery) (*mcp.CallToolResult, workerFitOutput, error) {
	out := workerFitOutput{}
	valid := map[string]bool{}
	for _, k := range []string{"strength", "agility", "one_handed", "two_handed", "polearm", "block", "archery", "harvest", "farm", "animal", "cook", "craft", "research", "labour"} {
		valid[k] = true
	}
	if !valid[q.Skill] {
		return nil, out, fmt.Errorf("skill must be an exact parser skill key")
	}
	limit := 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > 100 {
		return nil, out, fmt.Errorf("limit must be 1 through 100")
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
		return nil, out, fmt.Errorf("malformed saved world: %w", err)
	}
	if world.NPCs == nil {
		return nil, out, fmt.Errorf("missing saved NPC evidence")
	}
	if len(world.NPCs) > 100000 {
		return nil, out, fmt.Errorf("saved NPC evidence exceeds bound")
	}
	out = workerFitOutput{Source: raw.Source, SnapshotID: world.SnapshotID, IngestedAt: world.IngestedAt, RetrievedAt: raw.RetrievedAt, Skill: q.Skill, Limit: limit, Items: []workerFitItem{}, Note: "Advisory score = observed selected skill level, descending; competition rank and tied_workers include all matched workers before limit. Stable ties retain save order without expressing preference. Only explicit is_player_npc=true records. Unknown skills are excluded, never zero. Observations retain original injuries, acquired traits, job priorities and equipment, including null versus empty. No trait causality or productivity inference. Latest ingested save; retrieved_at is not save time."}
	items := []workerFitItem{}
	for _, n := range world.NPCs {
		if n["is_player_npc"] != true {
			continue
		}
		out.Total++
		v, present := n["skills"]
		if !present || v == nil {
			out.Unknown++
			continue
		}
		skills, ok := v.(map[string]any)
		if !ok {
			out.Malformed++
			continue
		}
		v, present = skills[q.Skill]
		if !present || v == nil {
			out.Unknown++
			continue
		}
		skill, ok := v.(map[string]any)
		if !ok {
			out.Malformed++
			continue
		}
		v, present = skill["level"]
		if !present || v == nil {
			out.Unknown++
			continue
		}
		score, ok := v.(float64)
		if !ok || score < 0 {
			out.Malformed++
			continue
		}
		items = append(items, workerFitItem{Score: score, Observations: n})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	for i := 0; i < len(items); {
		j := i + 1
		for j < len(items) && items[j].Score == items[i].Score {
			j++
		}
		for k := i; k < j; k++ {
			items[k].Rank = i + 1
			items[k].Tied = j - i
		}
		i = j
	}
	out.Matched = len(items)
	if len(items) > limit {
		items = items[:limit]
	}
	out.Items = items
	out.Returned = len(items)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
