// Package foodcontrol reads bounded live food observations. It never reads saves.
package foodcontrol

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Observation struct {
	Source             string           `json:"source"`
	SessionID          string           `json:"session_id"`
	CapturedAt         string           `json:"captured_at"`
	Available          bool             `json:"available"`
	Status             string           `json:"status"`
	WriteAvailable     bool             `json:"write_available"`
	WriteStatus        string           `json:"write_status"`
	Host               bool             `json:"host"`
	Player             string           `json:"player"`
	PlayerData         string           `json:"player_data"`
	CurrentSettlement  string           `json:"current_settlement"`
	Buildings          []map[string]any `json:"buildings"`
	Settlements        []map[string]any `json:"settlements"`
	Containers         []map[string]any `json:"containers"`
	ContainerDiscovery map[string]any   `json:"container_discovery,omitempty"`
	Discovery          []map[string]any `json:"discovery"`
	Scope              string           `json:"scope"`
	BudgetTruncated    bool             `json:"budget_truncated"`
}

type Result struct {
	Source         string       `json:"source"`
	Available      bool         `json:"available"`
	Status         string       `json:"status"`
	AgeSeconds     *float64     `json:"age_seconds"`
	Stale          bool         `json:"stale"`
	WriteAvailable bool         `json:"write_available"`
	Data           *Observation `json:"data"`
	Note           string       `json:"note"`
}

// Read refuses to infer missing fields, ownership, recruitment or consumption.
// Only the last 8MiB of the configured private runtime log is examined.
func Read(path string, now time.Time) (Result, error) {
	out := Result{Source: "ue4ss_live", Status: "awaiting_food_probe", Stale: true, Note: "Live food observations only; no save fallback. Residence does not establish recruitment, and storage rules are distinct from production targets. Writes remain disabled until the game setter and ownership are verified."}
	f, err := os.Open(path)
	if err != nil {
		out.Status = "log_unavailable"
		return out, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return out, err
	}
	start := st.Size() - 8*1024*1024
	if start < 0 {
		start = 0
	}
	if _, err = f.Seek(start, io.SeekStart); err != nil {
		return out, err
	}
	b, err := io.ReadAll(io.LimitReader(f, 8*1024*1024))
	if err != nil {
		return out, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 4096), 65536)
	var latest time.Time
	for scanner.Scan() {
		line := scanner.Text()
		at := strings.Index(line, "BWFOOD ")
		if at < 0 {
			continue
		}
		var v Observation
		if json.Unmarshal([]byte(line[at+7:]), &v) != nil || v.Source != "ue4ss_live" || v.SessionID == "" {
			continue
		}
		stamp, e := time.Parse(time.RFC3339, v.CapturedAt)
		if e != nil || stamp.After(now.Add(5*time.Second)) || stamp.Before(latest) {
			continue
		}
		latest = stamp
		copy := v
		copy.WriteAvailable = false
		copy.WriteStatus = "setter_unverified"
		out.Data = &copy
	}
	if err = scanner.Err(); err != nil {
		return out, fmt.Errorf("food log scan: %w", err)
	}
	if out.Data != nil {
		age := now.Sub(latest).Seconds()
		if age < 0 {
			age = 0
		}
		out.AgeSeconds = &age
		out.Stale = age > 20
		out.Status = out.Data.Status
		out.Available = out.Data.Available && !out.Stale
		if out.Stale {
			out.Status = "stale"
		}
	}
	// No writer is implemented: never trust a log's claim that writing is enabled.
	out.WriteAvailable = false
	return out, nil
}
