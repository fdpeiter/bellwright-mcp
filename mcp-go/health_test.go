package main

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func healthFixture(t *testing.T, values map[string]string) *backend {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, ok := values[r.URL.Path]
		if !ok {
			http.Error(w, "private C:/save/path", 503)
			return
		}
		w.Write([]byte(v))
	}))
	t.Cleanup(s.Close)
	b, e := newBackend(s.URL)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestHealthIndependentEvidence(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world/version": `{"snapshot_id":7,"ingested_at":"old-save","path":"private"}`, "/api/telemetry/latest": `{"received_at":123,"age_ms":100000,"data":{"session_id":"telemetry-session","captured_at":"capture"}}`, "/api/map": `{"status":{"session_id":"runtime-session","captured_at":"old-capture","age_seconds":200,"stale":true,"waiting_for_player":false,"message":"private"}}`})
	o := healthSummary(context.Background(), b, b)
	if *o.Sources[0].SnapshotID != 7 || o.Sources[0].Freshness != "unknown" || o.Sources[1].Freshness != "unknown" || *o.Sources[1].AgeMS != 100000 || o.Sources[2].Freshness != "stale" || *o.Sources[2].SessionID != "runtime-session" {
		t.Fatalf("%+v", o)
	}
	j, _ := json.Marshal(o)
	if strings.Contains(string(j), "private") {
		t.Fatal("private backend detail leaked")
	}
}
func TestHealthMissingAndUnavailable(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world/version": `{"snapshot_id":null}`, "/api/telemetry/latest": `{"data":null}`})
	o := healthSummary(context.Background(), b, b)
	if o.Sources[0].Availability != "no_data" || o.Sources[1].Availability != "no_data" || o.Sources[2].Availability != "unavailable" || o.Sources[2].AgeSeconds != nil {
		t.Fatalf("%+v", o)
	}
}
func TestHealthMCP(t *testing.T) {
	b := healthFixture(t, map[string]string{"/api/world/version": `{"snapshot_id":null}`, "/api/telemetry/latest": `{"data":null}`, "/api/map": `{"status":{"captured_at":"","session_id":"","age_seconds":null,"stale":true}}`})
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{Name: "health-test", Version: "1"}, nil)
	registerHealthTools(s, b, b)
	st, ct := mcp.NewInMemoryTransports()
	ss, e := s.Connect(ctx, st, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer ss.Close()
	cs, e := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, ct, nil)
	if e != nil {
		t.Fatal(e)
	}
	defer cs.Close()
	list, e := cs.ListTools(ctx, nil)
	if e != nil || len(list.Tools) != 1 || list.Tools[0].OutputSchema == nil {
		t.Fatalf("schema: %v %v", list, e)
	}
	r, e := cs.CallTool(ctx, &mcp.CallToolParams{Name: "read_connection_health", Arguments: map[string]any{}})
	if e != nil || r.IsError {
		t.Fatalf("call: %v %v", r, e)
	}
}

func TestHealthRejectsNonObjectTelemetry(t *testing.T) {
	for _, value := range []string{`{"data":[]}`, `{"data":42}`, `{"data":"invalid"}`} {
		b := healthFixture(t, map[string]string{"/api/telemetry/latest": value})
		o := healthSummary(context.Background(), b, b)
		if o.Sources[1].Availability != "unavailable" || o.Sources[1].Freshness != "unknown" {
			t.Fatalf("malformed telemetry became available: %+v", o.Sources[1])
		}
	}
}
