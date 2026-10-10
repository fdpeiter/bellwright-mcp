package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/url"
	"regexp"
)

var itemIDPattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type itemHistoryQuery struct {
	ItemID string `json:"item_id"`
	Limit  *int   `json:"limit,omitempty"`
}
type itemHistoryPoint struct {
	SnapshotID       *int     `json:"snapshot_id"`
	IngestedAt       string   `json:"ingested_at"`
	ObservedQty      *float64 `json:"observed_qty"`
	ObservedPlaytime *float64 `json:"observed_playtime"`
	SaveName         *string  `json:"save_name"`
	Region           *string  `json:"region"`
	Character        *string  `json:"character"`
	Map              *string  `json:"map"`
}
type itemHistoryOutput struct {
	Source         string             `json:"source"`
	RetrievedAt    string             `json:"retrieved_at"`
	ItemID         string             `json:"item_id"`
	TotalSnapshots *int               `json:"total_snapshots"`
	SampledPoints  *int               `json:"sampled_points"`
	Stride         *int               `json:"stride"`
	Sampled        *bool              `json:"sampled"`
	Matched        int                `json:"matched"`
	Returned       int                `json:"returned"`
	Truncated      int                `json:"truncated"`
	Limit          int                `json:"limit"`
	Points         []itemHistoryPoint `json:"points"`
	Note           string             `json:"note"`
}

func registerItemHistoryTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "read_item_history", Description: "Bounded saved quantity history for an exact item class ID; observed values and lineage may be unknown. Backend samples history before this limit.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}, InputSchema: map[string]any{"type": "object", "additionalProperties": false, "required": []string{"item_id"}, "properties": map[string]any{"item_id": map[string]any{"type": "string", "minLength": 1, "maxLength": 128, "pattern": "^[A-Za-z0-9_]+$"}, "limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 100, "default": 25}}}}, b.itemHistory)
}
func (b *backend) itemHistory(ctx context.Context, _ *mcp.CallToolRequest, q itemHistoryQuery) (*mcp.CallToolResult, itemHistoryOutput, error) {
	var o itemHistoryOutput
	limit, e := (spatialQuery{Limit: q.Limit}).validate()
	if e != nil {
		return nil, o, e
	}
	if len(q.ItemID) > 128 || !itemIDPattern.MatchString(q.ItemID) {
		return nil, o, fmt.Errorf("item_id must be 1 to 128 ASCII letters, digits or underscores")
	}
	_, raw, e := b.read(ctx, "/api/trends/item?"+url.Values{"name": {q.ItemID}}.Encode(), "companion_save_history", "")
	if e != nil {
		return nil, o, e
	}
	var d struct {
		Name           string             `json:"name"`
		Points         []itemHistoryPoint `json:"points"`
		TotalSnapshots *int               `json:"total_snapshots"`
		SampledPoints  *int               `json:"sampled_points"`
		Stride         *int               `json:"stride"`
		Sampled        *bool              `json:"sampled"`
	}
	if e = decodeEvidence(raw.Data, &d); e != nil {
		return nil, o, e
	}
	if d.Name != q.ItemID || d.Points == nil {
		return nil, o, fmt.Errorf("invalid item history evidence")
	}
	for _, p := range d.Points {
		if p.IngestedAt == "" || p.ObservedQty != nil && *p.ObservedQty < 0 || p.ObservedPlaytime != nil && *p.ObservedPlaytime < 0 {
			return nil, o, fmt.Errorf("invalid history point evidence")
		}
	}
	n := len(d.Points)
	if n > limit {
		n = limit
	}
	points := []itemHistoryPoint{}
	points = append(points, d.Points[len(d.Points)-n:]...)
	o = itemHistoryOutput{Source: raw.Source, RetrievedAt: raw.RetrievedAt, ItemID: d.Name, TotalSnapshots: d.TotalSnapshots, SampledPoints: d.SampledPoints, Stride: d.Stride, Sampled: d.Sampled, Matched: len(d.Points), Returned: n, Truncated: len(d.Points) - n, Limit: limit, Points: points, Note: "Saved history, not live stock or measured production. Returns latest bounded backend points in chronological order. Backend samples toward 240 points and retains the final snapshot (legacy algorithm can yield 241). matched/truncated cover backend sampled points, not all snapshots. Null observed quantity/playtime and lineage mean unknown; legacy numeric qty/playtime coerce missing values to zero and are intentionally not treated as observations. Ingestion timestamps are preserved, not save capture times. History can span saves, characters, regions and maps; compare lineage before inferring economic trends."}
	return nil, o, nil
}
