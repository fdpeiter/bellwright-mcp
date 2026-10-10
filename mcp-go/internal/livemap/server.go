// Package livemap serves continuous UE4SS telemetry only. Saves are never read.
package livemap

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

//go:embed index.html
var assets embed.FS

const MaxLogBytes int64 = 4 * 1024 * 1024
const MaxActors = 256

type Point struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Class      string          `json:"class"`
	Kind       string          `json:"kind"`
	Position   [3]float64      `json:"position"`
	Evidence   string          `json:"evidence"`
	CapturedAt string          `json:"captured_at"`
	AgeSeconds float64         `json:"age_seconds"`
	Stale      bool            `json:"stale"`
	Details    json.RawMessage `json:"details"`
}
type Status struct {
	SessionID         string   `json:"session_id"`
	CapturedAt        string   `json:"captured_at"`
	AgeSeconds        *float64 `json:"age_seconds"`
	Stale             bool     `json:"stale"`
	Waiting           bool     `json:"waiting_for_player"`
	Actors            int      `json:"actors"`
	Buildings         int      `json:"buildings"`
	Villagers         int      `json:"villagers"`
	SkippedOlderBytes int64    `json:"skipped_older_bytes"`
	InvalidRecords    int      `json:"invalid_records"`
	PlayerStatus      string   `json:"player_status"`
	Message           string   `json:"message"`
}
type Data struct {
	Status   Status   `json:"status"`
	Points   []Point  `json:"points"`
	Unknowns []string `json:"unknowns"`
}
type Server struct{ LogPath, BackgroundPath, CalibrationPath, ResourcePath, PlacementPath string }
type envelope struct {
	Version string          `json:"protocol_version"`
	Source  string          `json:"source"`
	Event   string          `json:"event"`
	Data    json.RawMessage `json:"data"`
}
type event struct {
	ScanID     string                      `json:"scan_id"`
	SessionID  string                      `json:"session_id"`
	CapturedAt string                      `json:"captured_at"`
	Sequence   *int64                      `json:"sequence"`
	ActorID    string                      `json:"actor_id"`
	QueryClass string                      `json:"query_class"`
	ClassName  string                      `json:"class_name"`
	Ownership  string                      `json:"ownership"`
	System     string                      `json:"coordinate_system"`
	Unit       string                      `json:"unit"`
	Position   *struct{ X, Y, Z *float64 } `json:"position"`
	Waiting    bool                        `json:"waiting_for_player"`
	Continuous bool                        `json:"continuous"`
}

func valid(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= 1e9 }

func actorLabel(query, id string) string {
	labels := map[string]string{"player": "Player", "TownNpc_C": "NPC", "PlayerShack_C": "Personal shack", "SharedChest_C": "Camp chest", "ShackChest_C": "Personal chest", "SmallFarm_C": "Farm", "Bloomery_C": "Bloomery", "Campfire_C": "Campfire", "ForagingCamp_C": "Foraging camp", "HousingTent_C": "Housing tent", "LoggingCamp_C": "Logging camp", "LumberjackSpot_C": "Lumberjack spot", "MiningCamp_C": "Mining camp", "MudCollector_C": "Mud collector", "Outhouse_C": "Outhouse", "PitSaw_C": "Pit saw", "RabbitTrap_C": "Rabbit trap", "ResearchDesk_C": "Research desk", "SimpleWorkbench_C": "Simple workbench", "SmokingRack_C": "Smoking rack", "Stockpile_C": "Stockpile", "TrapperCamp_C": "Trapper camp", "WaterCollector_C": "Water collector", "WeaverLoom_C": "Weaver loom", "WeaverTable_C": "Weaver table", "FastTravelSign_C": "Travel sign"}
	name := labels[query]
	if name == "" {
		name = strings.TrimSuffix(query, "_C")
	}
	if query == "player" {
		return name
	}
	suffix := id
	if at := strings.LastIndexAny(suffix, "./:"); at >= 0 {
		suffix = suffix[at+1:]
	}
	if len(suffix) > 16 {
		suffix = suffix[len(suffix)-16:]
	}
	return name + " · " + suffix
}
func tail(path string) ([]byte, int64, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, 0, e
	}
	defer f.Close()
	info, e := f.Stat()
	if e != nil {
		return nil, 0, e
	}
	if !info.Mode().IsRegular() {
		return nil, 0, fmt.Errorf("log must be regular file")
	}
	offset := info.Size() - MaxLogBytes
	if offset < 0 {
		offset = 0
	}
	b := make([]byte, info.Size()-offset)
	n, e := f.ReadAt(b, offset)
	if e != nil && e != io.EOF {
		return nil, offset, e
	}
	b = b[:n]
	if offset > 0 {
		at := bytes.IndexByte(b, '\n')
		if at < 0 {
			return nil, offset, nil
		}
		offset += int64(at + 1)
		b = b[at+1:]
	}
	if len(b) > 0 && b[len(b)-1] != '\n' {
		at := bytes.LastIndexByte(b, '\n')
		if at < 0 {
			b = nil
		} else {
			b = b[:at+1]
		}
	}
	return b, offset, nil
}
func (s Server) Data() (Data, error) {
	d := Data{Points: []Point{}, Status: Status{Stale: true, PlayerStatus: "Player telemetry unavailable", Message: "Waiting for continuous UE4SS telemetry; saves and bounded discovery are not used."}, Unknowns: []string{"All runtime actor ownership and settlement membership are unknown.", "Coordinates are engine units, unverified. Terrain, entrances, paths, inventory, resources and research are unavailable.", "Actor IDs are stable only within the current runtime session. Cached positions may be stale or actors may have despawned.", "Live probe has bounded cache and discovery radius; missing actors do not prove absence."}}
	b, skip, e := tail(s.LogPath)
	if e != nil {
		return d, e
	}
	d.Status.SkippedOlderBytes = skip
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 4096), 64*1024)
	latest := map[string]Point{}
	sequences := map[string]int64{}
	var session string
	var newest time.Time
	now := time.Now().UTC()
	for scanner.Scan() {
		line := scanner.Text()
		at := strings.Index(line, "BWADVISOR ")
		if at < 0 {
			continue
		}
		var env envelope
		if json.Unmarshal([]byte(line[at+10:]), &env) != nil {
			d.Status.InvalidRecords++
			continue
		}
		if env.Version != "0.3" || env.Source != "ue4ss_live" {
			continue
		}
		var ev event
		if json.Unmarshal(env.Data, &ev) != nil || ev.ScanID == "" || ev.SessionID != ev.ScanID || len(ev.ScanID) > 2048 || ev.Sequence == nil || *ev.Sequence < 0 {
			d.Status.InvalidRecords++
			continue
		}
		captured, e := time.Parse(time.RFC3339, ev.CapturedAt)
		if e != nil || captured.After(now.Add(5*time.Second)) {
			d.Status.InvalidRecords++
			continue
		}
		if env.Event != "actor" && env.Event != "heartbeat" && env.Event != "scan_start" {
			continue
		}
		if session != ev.ScanID {
			session = ev.ScanID
			latest = map[string]Point{}
			sequences = map[string]int64{}
			newest = time.Time{}
			d.Status.Waiting = false
		}
		if captured.After(newest) {
			newest = captured
		}
		if env.Event == "heartbeat" {
			d.Status.Waiting = ev.Waiting
			if ev.Waiting {
				latest = map[string]Point{}
				sequences = map[string]int64{}
			}
			continue
		}
		if env.Event != "actor" {
			continue
		}
		if ev.ActorID == "" || len(ev.ActorID) > 2048 || ev.Ownership != "unknown" || ev.System != "unreal_world" || ev.Unit != "engine_units_unverified" || ev.Position == nil || ev.Position.X == nil || ev.Position.Y == nil || ev.Position.Z == nil || !valid(*ev.Position.X) || !valid(*ev.Position.Y) || !valid(*ev.Position.Z) {
			d.Status.InvalidRecords++
			continue
		}
		if seq, ok := sequences[ev.ActorID]; ok && *ev.Sequence <= seq {
			continue
		}
		if _, ok := latest[ev.ActorID]; !ok && len(latest) >= MaxActors {
			continue
		}
		kind := "building"
		if ev.QueryClass == "TownNpc_C" {
			kind = "villager"
		}
		if ev.QueryClass == "player" {
			kind = "player"
		}
		if ev.QueryClass == "Actor" {
			kind = "runtime"
		}
		age := math.Max(0, now.Sub(captured).Seconds())
		latest[ev.ActorID] = Point{ID: "runtime:" + session + ":" + ev.ActorID, Name: actorLabel(ev.QueryClass, ev.ActorID), Class: ev.ClassName, Kind: kind, Position: [3]float64{*ev.Position.X, *ev.Position.Y, *ev.Position.Z}, Evidence: "Continuous UE4SS actor position; ownership unknown", CapturedAt: ev.CapturedAt, AgeSeconds: age, Stale: age > 45, Details: env.Data}
		sequences[ev.ActorID] = *ev.Sequence
	}
	if e = scanner.Err(); e != nil {
		return d, fmt.Errorf("log line exceeds bound")
	}
	d.Status.SessionID = session
	if !newest.IsZero() {
		d.Status.CapturedAt = newest.Format(time.RFC3339)
		age := math.Max(0, now.Sub(newest).Seconds())
		d.Status.AgeSeconds = &age
		d.Status.Stale = age > 10
		d.Status.Message = "Continuous runtime feed; event UTC timestamps determine freshness."
	}
	for _, p := range latest {
		d.Points = append(d.Points, p)
		if p.Kind == "building" {
			d.Status.Buildings++
		}
		if p.Kind == "villager" {
			d.Status.Villagers++
		}
		if p.Kind == "player" {
			d.Status.PlayerStatus = "Player runtime position available"
		}
	}
	sort.Slice(d.Points, func(i, j int) bool { return d.Points[i].ID < d.Points[j].ID })
	d.Status.Actors = len(d.Points)
	return d, nil
}
func allowedHost(host string) bool {
	h, _, e := net.SplitHostPort(host)
	if e != nil {
		h = host
	}
	return h == "127.0.0.1" || h == "localhost" || h == "::1"
}
func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/placements", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}
		ps, e := s.Placements()
		if e != nil {
			http.Error(w, "Placement candidates unavailable", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ps)
	})
	mux.HandleFunc("/api/resources", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}
		b, e := s.Resources()
		if e != nil {
			http.Error(w, "Resource baseline unavailable or invalid", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(b)
	})
	mux.HandleFunc("/api/background", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}
		b, e := s.Background()
		if e != nil {
			http.Error(w, "Background unavailable or invalid", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(b)
	})
	mux.HandleFunc("/map-background.png", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}
		if _, e := s.Background(); e != nil {
			http.Error(w, "Background unavailable or invalid", 404)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "image/png")
		http.ServeFile(w, r, s.BackgroundPath)
	})
	mux.HandleFunc("/api/map", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "GET only", 405)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json")
		d, e := s.Data()
		if e != nil {
			w.WriteHeader(503)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Continuous UE4SS log unavailable or invalid. No save fallback."})
			return
		}
		_ = json.NewEncoder(w).Encode(d)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; img-src 'self' data:")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		b, _ := assets.ReadFile("index.html")
		_, _ = w.Write(b)
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowedHost(r.Host) {
			http.Error(w, "loopback host required", 403)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
