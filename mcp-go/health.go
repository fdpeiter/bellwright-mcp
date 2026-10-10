package main

import (
	"context"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"sync"
)

type healthEvidence struct {
	Source           string   `json:"source"`
	RetrievedAt      *string  `json:"retrieved_at"`
	Availability     string   `json:"availability"`
	Freshness        string   `json:"freshness"`
	SnapshotID       *int64   `json:"snapshot_id"`
	IngestedAt       *string  `json:"ingested_at"`
	CapturedAt       *string  `json:"captured_at"`
	SessionID        *string  `json:"session_id"`
	AgeMS            *float64 `json:"age_ms"`
	AgeSeconds       *float64 `json:"age_seconds"`
	ReceivedAt       *float64 `json:"received_at"`
	Stale            *bool    `json:"stale"`
	WaitingForPlayer *bool    `json:"waiting_for_player"`
}
type healthOutput struct {
	Sources []healthEvidence `json:"sources"`
	Note    string           `json:"note"`
}

func registerHealthTools(s *mcp.Server, companion, live *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "read_connection_health", Description: "Independent save, companion telemetry and continuous live capture health. Source timestamps only; no save fallback.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, func(ctx context.Context, _ *mcp.CallToolRequest, _ input) (*mcp.CallToolResult, healthOutput, error) {
		return nil, healthSummary(ctx, companion, live), nil
	})
}
func healthSummary(ctx context.Context, companion, live *backend) healthOutput {
	out := healthOutput{Sources: make([]healthEvidence, 3), Note: "Save ingestion time and telemetry receipt time are not game capture time. Unknown ages remain null. Live runtime ownership is unknown; no save fallback."}
	var wg sync.WaitGroup
	for i, spec := range []struct {
		b            *backend
		path, source string
	}{{companion, "/api/world/version", "companion_save"}, {companion, "/api/telemetry/latest", "companion_telemetry"}, {live, "/api/map", "ue4ss_live"}} {
		wg.Add(1)
		go func(i int, b *backend, path, source string) {
			defer wg.Done()
			h := healthEvidence{Source: source, Availability: "unavailable", Freshness: "unknown"}
			defer func() { out.Sources[i] = h }()
			if b == nil {
				return
			}
			_, raw, err := b.read(ctx, path, source, "")
			if err != nil {
				return
			}
			m, ok := raw.Data.(map[string]any)
			if !ok {
				return
			}
			h.RetrievedAt = &raw.RetrievedAt
			h.Availability = "available"
			switch source {
			case "companion_save":
				if _, present := m["snapshot_id"]; !present {
					h.Availability = "unavailable"
					return
				}
				var v struct {
					ID       *int64  `json:"snapshot_id"`
					Ingested *string `json:"ingested_at"`
				}
				if decodeEvidence(m, &v) != nil {
					h.Availability = "unavailable"
					return
				}
				h.SnapshotID = v.ID
				h.IngestedAt = v.Ingested
				if v.ID == nil {
					h.Availability = "no_data"
				}
			case "companion_telemetry":
				if _, present := m["data"]; !present {
					h.Availability = "unavailable"
					return
				}
				var v struct {
					Age      *float64 `json:"age_ms"`
					Received *float64 `json:"received_at"`
				}
				if decodeEvidence(m, &v) != nil {
					h.Availability = "unavailable"
					return
				}
				h.AgeMS = v.Age
				h.ReceivedAt = v.Received
				if m["data"] == nil {
					h.Availability = "no_data"
					return
				}
				if d, ok := m["data"].(map[string]any); ok {
					var capture struct {
						At      *string `json:"captured_at"`
						Session *string `json:"session_id"`
					}
					if decodeEvidence(d, &capture) == nil {
						h.CapturedAt = capture.At
						h.SessionID = capture.Session
					}
				} else {
					h.Availability = "unavailable"
				}
			case "ue4ss_live":
				d, ok := m["status"].(map[string]any)
				if !ok {
					h.Availability = "unavailable"
					return
				}
				if _, present := d["captured_at"]; !present {
					h.Availability = "unavailable"
					return
				}
				var v struct {
					At      *string  `json:"captured_at"`
					Session *string  `json:"session_id"`
					Age     *float64 `json:"age_seconds"`
					Stale   *bool    `json:"stale"`
					Waiting *bool    `json:"waiting_for_player"`
				}
				if decodeEvidence(d, &v) != nil {
					h.Availability = "unavailable"
					return
				}
				h.CapturedAt = v.At
				h.SessionID = v.Session
				h.AgeSeconds = v.Age
				h.Stale = v.Stale
				h.WaitingForPlayer = v.Waiting
				if v.At == nil || *v.At == "" {
					h.CapturedAt = nil
					h.Availability = "no_data"
				}
				if v.Session != nil && *v.Session == "" {
					h.SessionID = nil
				}
				if h.Availability == "available" && v.Stale != nil {
					if *v.Stale {
						h.Freshness = "stale"
					} else if v.Age != nil {
						h.Freshness = "fresh"
					}
				}
			}
			if h.AgeMS != nil && *h.AgeMS < 0 {
				h.AgeMS = nil
			}
			if h.AgeSeconds != nil && *h.AgeSeconds < 0 {
				h.AgeSeconds = nil
				h.Freshness = "unknown"
			}
		}(i, spec.b, spec.path, spec.source)
	}
	wg.Wait()
	return out
}
