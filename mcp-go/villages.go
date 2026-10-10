package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"strings"
)

type villageQuery struct {
	Name  string `json:"name,omitempty"`
	Limit *int   `json:"limit,omitempty"`
}
type villageProgress struct {
	Name       string   `json:"name"`
	Trust      *float64 `json:"trust"`
	TrustLevel *int     `json:"trust_level"`
	Prosperity *float64 `json:"prosperity"`
	Liberated  *bool    `json:"liberated"`
}
type villageOutput struct {
	Source     string            `json:"source"`
	SnapshotID *int64            `json:"snapshot_id"`
	IngestedAt *string           `json:"ingested_at"`
	State      string            `json:"state"`
	Total      int               `json:"total"`
	Matched    int               `json:"matched"`
	Returned   int               `json:"returned"`
	Truncated  int               `json:"truncated"`
	Villages   []villageProgress `json:"villages"`
	Note       string            `json:"note"`
}

func registerVillageTools(s *mcp.Server, b *backend) {
	schema := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"name": map[string]any{"type": "string", "maxLength": 128}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 25}}}
	mcp.AddTool(s, &mcp.Tool{Name: "read_saved_village_progress", InputSchema: schema, Description: "Bounded case-insensitive village name search over saved trust, prosperity and liberation evidence. Missing fields remain unknown.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.villageProgress)
}
func (b *backend) villageProgress(ctx context.Context, _ *mcp.CallToolRequest, q villageQuery) (*mcp.CallToolResult, villageOutput, error) {
	o := villageOutput{Source: "companion_save", Villages: []villageProgress{}, Note: "Latest ingested save only; ingestion is not game capture time. Missing village evidence or fields does not establish zero trust, zero prosperity or unliberated status. No live fallback."}
	limit := 25
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > 100 || len(q.Name) > 128 {
		return nil, o, fmt.Errorf("limit must be 1 through 100 and name at most 128 bytes")
	}
	_, raw, e := b.read(ctx, "/api/world", "companion_save", "")
	if e != nil {
		return nil, o, fmt.Errorf("saved world unavailable")
	}
	var w struct {
		SnapshotID *int64             `json:"snapshot_id"`
		IngestedAt *string            `json:"ingested_at"`
		Villages   *[]villageProgress `json:"villages"`
	}
	if decodeEvidence(raw.Data, &w) != nil {
		return nil, o, fmt.Errorf("invalid village evidence")
	}
	o.SnapshotID = w.SnapshotID
	o.IngestedAt = w.IngestedAt
	if w.Villages == nil {
		o.State = "missing"
		return nil, o, nil
	}
	o.State = "available"
	o.Total = len(*w.Villages)
	if o.Total == 0 {
		o.State = "empty"
	}
	for _, v := range *w.Villages {
		if v.Name == "" {
			return nil, villageOutput{}, fmt.Errorf("missing village identity")
		}
		if q.Name != "" && !strings.Contains(strings.ToLower(v.Name), strings.ToLower(q.Name)) {
			continue
		}
		o.Matched++
		if len(o.Villages) < limit {
			o.Villages = append(o.Villages, v)
		}
	}
	o.Returned = len(o.Villages)
	o.Truncated = o.Matched - o.Returned
	return nil, o, nil
}
