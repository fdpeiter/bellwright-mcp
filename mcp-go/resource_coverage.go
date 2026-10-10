package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"sort"
)

type coverageQuery struct {
	Category *string `json:"category,omitempty" jsonschema:"Exact category; omitted selects all, empty selects unknown category"`
	Limit    *int    `json:"limit,omitempty" jsonschema:"Maximum categories, default 25, maximum 100"`
}
type coverageBounds struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MinZ float64 `json:"min_z"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
	MaxZ float64 `json:"max_z"`
}
type coverageCategory struct {
	Category      string         `json:"category"`
	CategoryKnown bool           `json:"category_known"`
	Nodes         int            `json:"nodes"`
	Bounds        coverageBounds `json:"bounds"`
}
type coverageOutput struct {
	Source          string             `json:"source"`
	RetrievedAt     string             `json:"retrieved_at"`
	Schema          string             `json:"backend_schema"`
	SourceKind      string             `json:"source_kind"`
	TotalNodes      int                `json:"total_nodes"`
	MatchedNodes    int                `json:"matched_nodes"`
	TotalCategories int                `json:"total_categories"`
	Matched         int                `json:"matched"`
	Returned        int                `json:"returned"`
	Truncated       int                `json:"truncated"`
	Limit           int                `json:"limit"`
	Categories      []coverageCategory `json:"categories"`
	CoordinateUnits string             `json:"coordinate_units"`
	Availability    string             `json:"availability"`
	Unknowns        []string           `json:"unknowns"`
	Note            string             `json:"note"`
}

func registerResourceCoverageTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "summarize_resource_coverage", Description: "Summarize loaded packaged resource categories and observed XYZ bounds in engine units. Counts are loaded evidence only; availability and productivity unknown. Exact optional category, including empty for unknown category; category limit default 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.resourceCoverage)
}
func (b *backend) resourceCoverage(ctx context.Context, _ *mcp.CallToolRequest, q coverageQuery) (*mcp.CallToolResult, coverageOutput, error) {
	var out coverageOutput
	limit, err := (spatialQuery{Limit: q.Limit}).validate()
	if err != nil {
		return nil, out, err
	}
	if q.Category != nil && len(*q.Category) > 128 {
		return nil, out, fmt.Errorf("category exceeds 128 bytes")
	}
	_, raw, err := b.read(ctx, "/api/resources", "packaged_game_assets", "")
	if err != nil {
		return nil, out, err
	}
	var baseline struct {
		Schema   string         `json:"schema"`
		Source   string         `json:"source_kind"`
		Nodes    []resourceItem `json:"nodes"`
		Unknowns []string       `json:"unknowns"`
	}
	if err = decodeEvidence(raw.Data, &baseline); err != nil {
		return nil, out, err
	}
	if baseline.Schema != "bellwright.harvestable-baseline.v1" || baseline.Source != "packaged_game_asset" || baseline.Nodes == nil || len(baseline.Nodes) > 100000 || len(baseline.Unknowns) == 0 {
		return nil, out, fmt.Errorf("invalid resource evidence or provenance")
	}
	categories := map[string]coverageCategory{}
	for _, n := range baseline.Nodes {
		if err = validatePosition(n.Position); err != nil {
			return nil, out, err
		}
		if n.ID == "" || n.Class == "" {
			return nil, out, fmt.Errorf("missing resource identity evidence")
		}
		c, exists := categories[n.Category]
		p := n.Position
		if !exists {
			c = coverageCategory{Category: n.Category, CategoryKnown: n.Category != "", Bounds: coverageBounds{p[0], p[1], p[2], p[0], p[1], p[2]}}
		}
		c.Nodes++
		c.Bounds.MinX = min(c.Bounds.MinX, p[0])
		c.Bounds.MinY = min(c.Bounds.MinY, p[1])
		c.Bounds.MinZ = min(c.Bounds.MinZ, p[2])
		c.Bounds.MaxX = max(c.Bounds.MaxX, p[0])
		c.Bounds.MaxY = max(c.Bounds.MaxY, p[1])
		c.Bounds.MaxZ = max(c.Bounds.MaxZ, p[2])
		categories[n.Category] = c
	}
	out = coverageOutput{Source: raw.Source, RetrievedAt: raw.RetrievedAt, Schema: baseline.Schema, SourceKind: baseline.Source, TotalNodes: len(baseline.Nodes), TotalCategories: len(categories), Limit: limit, Categories: []coverageCategory{}, CoordinateUnits: "engine", Availability: "unknown", Unknowns: baseline.Unknowns, Note: "Loaded packaged static nodes only; no world census, measured productivity or current availability. Bounds span observed XYZ coordinates; retrieval time is not capture time. Empty category identifies unknown category evidence. Categories sorted by exact category."}
	keys := make([]string, 0, len(categories))
	for k := range categories {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if q.Category != nil && k != *q.Category {
			continue
		}
		c := categories[k]
		out.Matched++
		out.MatchedNodes += c.Nodes
		if len(out.Categories) < limit {
			out.Categories = append(out.Categories, c)
		}
	}
	out.Returned = len(out.Categories)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
