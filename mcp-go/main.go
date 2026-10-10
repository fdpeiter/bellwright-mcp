package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type backend struct {
	base   string
	client *http.Client
}
type input struct{}
type output struct {
	Source      string `json:"source"`
	RetrievedAt string `json:"retrieved_at"`
	Data        any    `json:"data"`
	Note        string `json:"note"`
}

func newBackend(raw string) (*backend, error) {
	u, e := url.Parse(raw)
	if e != nil {
		return nil, e
	}
	if u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, fmt.Errorf("use an HTTP loopback origin")
	}
	if h := u.Hostname(); h != "127.0.0.1" && h != "localhost" && h != "::1" {
		return nil, fmt.Errorf("backend must be loopback")
	}
	return &backend{strings.TrimRight(raw, "/"), &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("redirect rejected") }}}, nil
}
func (b *backend) read(ctx context.Context, path, source, note string) (*mcp.CallToolResult, output, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, b.base+path, nil)
	if e != nil {
		return nil, output{}, e
	}
	r, e := b.client.Do(req)
	if e != nil {
		return nil, output{}, fmt.Errorf("backend unavailable: %w", e)
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return nil, output{}, fmt.Errorf("backend returned HTTP %d", r.StatusCode)
	}
	data, e := io.ReadAll(io.LimitReader(r.Body, 32*1024*1024+1))
	if e != nil {
		return nil, output{}, e
	}
	if len(data) > 32*1024*1024 {
		return nil, output{}, fmt.Errorf("response exceeds 32MiB")
	}
	if !json.Valid(data) {
		return nil, output{}, fmt.Errorf("backend returned invalid JSON")
	}
	var decoded any
	if err := json.Unmarshal(data, &decoded); err != nil {
		return nil, output{}, err
	}
	return nil, output{source, time.Now().UTC().Format(time.RFC3339), decoded, note}, nil
}
func server(companion, live *backend) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "bellwright-steward-go", Version: "0.4.0"}, nil)
	specs := []struct {
		name, path, source, note string
		b                        *backend
	}{
		{"read_saved_world", "/api/world", "companion_save", "Latest ingested save; not live game state.", companion},
		{"list_save_snapshots", "/api/snapshots", "companion_save_history", "Save metadata and ingestion times.", companion},
		{"read_save_trends", "/api/trends", "companion_save_history", "Rates inferred from saves, not live production counters.", companion},
		{"read_companion_telemetry", "/api/telemetry/latest", "companion_telemetry", "Inspect age_ms and data; missing data means no telemetry received.", companion},
		{"read_live_map", "/api/map", "ue4ss_live", "Inspect status.stale and capture ages; runtime ownership unknown. No save fallback.", live},
	}
	for _, spec := range specs {
		p := spec
		mcp.AddTool(s, &mcp.Tool{Name: p.name, Description: p.note, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, output, error) {
			return p.b.read(ctx, p.path, p.source, p.note)
		})
	}
	mcp.AddTool(s, &mcp.Tool{Name: "read_resource_baseline", InputSchema: boundedInputSchema("category", 128), Description: "Bounded packaged nodes filtered by category and XY bounds/radius in engine units. Availability unknown; counts cover only the loaded baseline.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, live.resources)
	mcp.AddTool(s, &mcp.Tool{Name: "read_foraging_candidates", InputSchema: boundedInputSchema("target_runtime_actor_id", 256), Description: "Bounded provisional placements filtered by target ID and practical-candidate XY bounds/radius in engine units. Coverage is not measured yield; availability/buildability unknown.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, live.placements)
	registerGearTools(s, companion)
	registerItemHistoryTools(s, companion)
	registerVillagerTools(s, companion)
	registerInventoryTools(s, companion)
	registerSquadTools(s, companion)
	registerVillageTools(s, companion)
	registerHealthTools(s, companion, live)
	registerWorkerFitTools(s, companion)
	registerJobPriorityTools(s, companion)
	registerHousingTools(s, companion)
	registerStockTargetTools(s, companion)
	registerStockRecommendationTools(s, companion)
	registerObservedItemTools(s, companion)
	registerResourceCoverageTools(s, live)
	return s
}

const defaultQueryLimit = 25
const maxQueryLimit = 100

func boundedInputSchema(filter string, maxLength int) map[string]any {
	number := func() map[string]any { return map[string]any{"type": "number", "minimum": -1e8, "maximum": 1e8} }
	return map[string]any{
		"type": "object", "additionalProperties": false,
		"properties": map[string]any{
			filter:   map[string]any{"type": "string", "maxLength": maxLength, "description": "Exact evidence match; omitted or empty means all"},
			"limit":  map[string]any{"type": "integer", "minimum": 1, "maximum": maxQueryLimit, "default": defaultQueryLimit},
			"center": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"x", "y"}, "properties": map[string]any{"x": number(), "y": number()}, "description": "XY center in engine units; requires radius"},
			"radius": map[string]any{"type": "number", "exclusiveMinimum": 0, "maximum": 1e8, "description": "XY engine units; requires center"},
			"bounds": map[string]any{"type": "object", "additionalProperties": false, "required": []string{"min_x", "min_y", "max_x", "max_y"}, "properties": map[string]any{"min_x": number(), "min_y": number(), "max_x": number(), "max_y": number()}, "description": "Inclusive XY engine bounds; intersects radius when both supplied"},
		},
	}
}

type xy struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type queryBounds struct {
	MinX float64 `json:"min_x"`
	MinY float64 `json:"min_y"`
	MaxX float64 `json:"max_x"`
	MaxY float64 `json:"max_y"`
}
type spatialQuery struct {
	Center *xy          `json:"center,omitempty" jsonschema:"XY center in engine units; requires radius"`
	Radius *float64     `json:"radius,omitempty" jsonschema:"Positive XY radius in engine units; requires center"`
	Bounds *queryBounds `json:"bounds,omitempty" jsonschema:"Inclusive XY bounds; intersects radius when both supplied"`
	Limit  *int         `json:"limit,omitempty" jsonschema:"Maximum items; default 25; allowed 1 through 100"`
}
type resourceQuery struct {
	spatialQuery
	Category string `json:"category,omitempty" jsonschema:"Exact baseline_category match; omitted means all categories"`
}
type placementQuery struct {
	spatialQuery
	TargetID string `json:"target_runtime_actor_id,omitempty" jsonschema:"Exact target ID match; runtime ownership remains unknown"`
}
type resourceItem struct {
	ID       string    `json:"id"`
	Category string    `json:"baseline_category"`
	Class    string    `json:"class_asset_path"`
	Position []float64 `json:"position"`
}
type placementSite struct {
	Position []float64 `json:"position"`
	Count    int       `json:"baseline_food_node_count"`
	Distance float64   `json:"distance_to_core_engine_units"`
}
type placementItem struct {
	Schema      string        `json:"schema"`
	ID          string        `json:"target_runtime_actor_id"`
	Radius      float64       `json:"radius_engine_units_assumed"`
	Current     placementSite `json:"current"`
	Practical   placementSite `json:"practical_candidate"`
	Assumptions []string      `json:"assumptions"`
}
type boundedOutput[T any] struct {
	Source       string   `json:"source"`
	RetrievedAt  string   `json:"retrieved_at"`
	Schema       string   `json:"backend_schema"`
	Total        int      `json:"total" jsonschema:"Items in loaded backend evidence; not a world census"`
	Matched      int      `json:"matched" jsonschema:"Items matching all filters before limit"`
	Returned     int      `json:"returned"`
	Truncated    int      `json:"truncated" jsonschema:"Matching items omitted by limit"`
	Limit        int      `json:"limit"`
	Items        []T      `json:"items"`
	Availability string   `json:"availability"`
	Unknowns     []string `json:"unknowns"`
	Note         string   `json:"note"`
}

func validCoordinate(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= 1e8
}
func (q spatialQuery) validate() (int, error) {
	limit := defaultQueryLimit
	if q.Limit != nil {
		limit = *q.Limit
	}
	if limit < 1 || limit > maxQueryLimit {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	if (q.Center == nil) != (q.Radius == nil) {
		return 0, fmt.Errorf("center and radius must be provided together")
	}
	if q.Center != nil && (!validCoordinate(q.Center.X) || !validCoordinate(q.Center.Y) || !validCoordinate(*q.Radius) || *q.Radius <= 0) {
		return 0, fmt.Errorf("invalid center or radius; use finite engine units and positive radius at most 1e8")
	}
	if b := q.Bounds; b != nil {
		if !validCoordinate(b.MinX) || !validCoordinate(b.MinY) || !validCoordinate(b.MaxX) || !validCoordinate(b.MaxY) || b.MinX > b.MaxX || b.MinY > b.MaxY {
			return 0, fmt.Errorf("invalid bounds; minimum must not exceed maximum")
		}
	}
	return limit, nil
}
func validatePosition(p []float64) error {
	if len(p) != 3 {
		return fmt.Errorf("backend position must have three observed coordinates")
	}
	for _, v := range p {
		if !validCoordinate(v) {
			return fmt.Errorf("invalid backend coordinate")
		}
	}
	return nil
}
func (q spatialQuery) matches(p []float64) bool {
	if b := q.Bounds; b != nil && (p[0] < b.MinX || p[0] > b.MaxX || p[1] < b.MinY || p[1] > b.MaxY) {
		return false
	}
	return q.Center == nil || math.Hypot(p[0]-q.Center.X, p[1]-q.Center.Y) <= *q.Radius
}
func decodeEvidence(data any, dst any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dst)
}
func (b *backend) resources(ctx context.Context, _ *mcp.CallToolRequest, q resourceQuery) (*mcp.CallToolResult, boundedOutput[resourceItem], error) {
	var out boundedOutput[resourceItem]
	limit, err := q.validate()
	if err != nil {
		return nil, out, err
	}
	if len(q.Category) > 128 {
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
	out = boundedOutput[resourceItem]{Source: raw.Source, RetrievedAt: raw.RetrievedAt, Schema: baseline.Schema, Total: len(baseline.Nodes), Limit: limit, Items: []resourceItem{}, Availability: "unknown", Unknowns: baseline.Unknowns, Note: "Packaged static positions only. Depletion, regrowth, seasonal availability and yield unknown. Counts cover loaded evidence only; retrieval time is not capture time. Backend order is preserved, not nearest-first."}
	for _, n := range baseline.Nodes {
		if err = validatePosition(n.Position); err != nil {
			return nil, boundedOutput[resourceItem]{}, err
		}
		if n.ID == "" || n.Class == "" {
			return nil, boundedOutput[resourceItem]{}, fmt.Errorf("missing resource identity evidence")
		}
		if (q.Category == "" || n.Category == q.Category) && q.matches(n.Position) {
			out.Matched++
			if len(out.Items) < limit {
				out.Items = append(out.Items, n)
			}
		}
	}
	out.Returned = len(out.Items)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
func (b *backend) placements(ctx context.Context, _ *mcp.CallToolRequest, q placementQuery) (*mcp.CallToolResult, boundedOutput[placementItem], error) {
	var out boundedOutput[placementItem]
	limit, err := q.validate()
	if err != nil {
		return nil, out, err
	}
	if len(q.TargetID) > 256 {
		return nil, out, fmt.Errorf("target ID exceeds 256 bytes")
	}
	_, raw, err := b.read(ctx, "/api/placements", "baseline_placement_analysis", "")
	if err != nil {
		return nil, out, err
	}
	var items []placementItem
	if err = decodeEvidence(raw.Data, &items); err != nil {
		return nil, out, err
	}
	if items == nil || len(items) > 256 {
		return nil, out, fmt.Errorf("invalid placement evidence or bound")
	}
	out = boundedOutput[placementItem]{Source: raw.Source, RetrievedAt: raw.RetrievedAt, Schema: "bellwright.foraging-placement-candidate.v1", Total: len(items), Limit: limit, Items: []placementItem{}, Availability: "unknown", Unknowns: []string{"Buildability, terrain, hauling access, work radius, current availability and measured yield are unverified.", "Runtime actor ownership and settlement membership are unknown."}, Note: "Filters use practical_candidate XY engine coordinates. Coverage counts are packaged node estimates, not output. Retrieval time is not capture time. Backend order is preserved."}
	for _, p := range items {
		if err = validatePosition(p.Current.Position); err != nil {
			return nil, boundedOutput[placementItem]{}, err
		}
		if err = validatePosition(p.Practical.Position); err != nil {
			return nil, boundedOutput[placementItem]{}, err
		}
		if p.Schema != out.Schema || p.ID == "" || len(p.Assumptions) == 0 || !validCoordinate(p.Radius) || p.Radius <= 0 || p.Radius > 1e6 || p.Current.Count < 0 || p.Practical.Count < 0 {
			return nil, boundedOutput[placementItem]{}, fmt.Errorf("invalid placement evidence")
		}
		if (q.TargetID == "" || q.TargetID == p.ID) && q.matches(p.Practical.Position) {
			out.Matched++
			if len(out.Items) < limit {
				out.Items = append(out.Items, p)
			}
		}
	}
	out.Returned = len(out.Items)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
func main() {
	c := flag.String("companion-url", "http://127.0.0.1:8710", "Companion loopback origin")
	l := flag.String("live-url", "http://127.0.0.1:18126", "UE4SS map loopback origin")
	flag.Parse()
	cb, e := newBackend(*c)
	if e != nil {
		log.Fatal(e)
	}
	lb, e := newBackend(*l)
	if e != nil {
		log.Fatal(e)
	}
	if e = server(cb, lb).Run(context.Background(), &mcp.StdioTransport{}); e != nil {
		log.Fatal(e)
	}
}
