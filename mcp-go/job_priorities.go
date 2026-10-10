package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"sort"
)

type jobPriorityQuery struct {
	Limit *int `json:"limit,omitempty"`
}
type jobPriorityItem struct {
	Observations map[string]any `json:"observations"`
	State        string         `json:"priority_evidence_state"`
	Strongest    *float64       `json:"strongest_observed_priority"`
	Jobs         []string       `json:"strongest_observed_jobs"`
	Conflict     bool           `json:"observed_tie"`
}
type jobPriorityOutput struct {
	Source      string            `json:"source"`
	SnapshotID  *int              `json:"snapshot_id"`
	IngestedAt  *string           `json:"ingested_at"`
	RetrievedAt string            `json:"retrieved_at"`
	Total       int               `json:"total"`
	Missing     int               `json:"unknown_priority_maps"`
	Malformed   int               `json:"malformed_priority_maps"`
	Empty       int               `json:"observed_empty_priority_maps"`
	Conflicts   int               `json:"observed_ties"`
	Returned    int               `json:"returned"`
	Truncated   int               `json:"truncated"`
	Limit       int               `json:"limit"`
	Items       []jobPriorityItem `json:"items"`
	Note        string            `json:"note"`
}

func registerJobPriorityTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "audit_saved_job_priorities", Description: "Advisory audit of recruited NPC saved job priorities: report ties at the strongest observed displayed priority. Missing jobs are unknown, not default. Empty maps never imply idle. Default limit 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.jobPriorityAudit)
}
func (b *backend) jobPriorityAudit(ctx context.Context, _ *mcp.CallToolRequest, q jobPriorityQuery) (*mcp.CallToolResult, jobPriorityOutput, error) {
	out := jobPriorityOutput{}
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
	out = jobPriorityOutput{Source: raw.Source, SnapshotID: world.SnapshotID, IngestedAt: world.IngestedAt, RetrievedAt: raw.RetrievedAt, Limit: limit, Items: []jobPriorityItem{}, Note: "Advisory rule: at least two observed jobs share the minimum displayed priority. Parser formula displayed=10-raw_weight; valid displayed range 0..9; lower numbers are stronger. Original displayed values remain in observations; raw weights are not supplied or fabricated. Only non-default jobs are stored, so omitted jobs remain unknown and no default is inserted. Ties describe potential competition, not proven scheduling conflicts. Empty or missing maps do not establish idle/non-worker status. Only explicit is_player_npc=true; save order preserved. Latest ingested save; retrieval time is not save time."}
	for _, n := range world.NPCs {
		if n["is_player_npc"] != true {
			continue
		}
		out.Total++
		item := jobPriorityItem{Observations: n, State: "observed", Jobs: []string{}}
		v, present := n["job_priorities"]
		if !present || v == nil {
			item.State = "unknown"
			out.Missing++
		} else if jobs, ok := v.(map[string]any); !ok {
			item.State = "malformed"
			out.Malformed++
		} else if len(jobs) == 0 {
			item.State = "observed_empty"
			out.Empty++
		} else {
			for job, v := range jobs {
				p, ok := v.(float64)
				if !ok || p < 0 || p > 9 || p != float64(int(p)) {
					item.State = "malformed"
					break
				}
				if item.Strongest == nil || p < *item.Strongest {
					score := p
					item.Strongest = &score
					item.Jobs = []string{job}
				} else if p == *item.Strongest {
					item.Jobs = append(item.Jobs, job)
				}
			}
			if item.State == "malformed" {
				out.Malformed++
				item.Strongest = nil
				item.Jobs = []string{}
			} else {
				sort.Strings(item.Jobs)
				item.Conflict = len(item.Jobs) > 1
				if item.Conflict {
					out.Conflicts++
				}
			}
		}
		if len(out.Items) < limit {
			out.Items = append(out.Items, item)
		}
	}
	out.Returned = len(out.Items)
	out.Truncated = out.Total - out.Returned
	return nil, out, nil
}
