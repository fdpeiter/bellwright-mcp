package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"io"
	"log"
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
	s := mcp.NewServer(&mcp.Implementation{Name: "bellwright-steward-go", Version: "0.1.0"}, nil)
	specs := []struct {
		name, path, source, note string
		b                        *backend
	}{
		{"read_saved_world", "/api/world", "companion_save", "Latest ingested save; not live game state.", companion},
		{"list_save_snapshots", "/api/snapshots", "companion_save_history", "Save metadata and ingestion times.", companion},
		{"read_save_trends", "/api/trends", "companion_save_history", "Rates inferred from saves, not live production counters.", companion},
		{"read_companion_telemetry", "/api/telemetry/latest", "companion_telemetry", "Inspect age_ms and data; missing data means no telemetry received.", companion},
		{"read_live_map", "/api/map", "ue4ss_live", "Inspect status.stale and capture ages; runtime ownership unknown. No save fallback.", live},
		{"read_resource_baseline", "/api/resources", "packaged_game_assets", "Baseline nodes; depletion, regrowth and current availability unknown.", live},
		{"read_foraging_candidates", "/api/placements", "baseline_placement_analysis", "Node coverage estimates; terrain, access, yields and buildability unverified.", live},
	}
	for _, spec := range specs {
		p := spec
		mcp.AddTool(s, &mcp.Tool{Name: p.name, Description: p.note, Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, output, error) {
			return p.b.read(ctx, p.path, p.source, p.note)
		})
	}
	return s
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
