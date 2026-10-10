package livemap

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fixture(x int, at string, source string) string {
	return fmt.Sprintf(`BWADVISOR {"protocol_version":"0.3","source":"%s","event":"actor","data":{"scan_id":"synthetic-session","session_id":"synthetic-session","captured_at":"%s","sequence":%d,"actor_id":"SyntheticActor","query_class":"TownNpc_C","class_name":"SyntheticNpc","ownership":"unknown","coordinate_system":"unreal_world","unit":"engine_units_unverified","position":{"x":%d,"y":200,"z":300}}}`+"\n", source, at, x, x)
}
func TestLiveFreshnessUpdateAndNoSave(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "UE4SS.log")
	at := time.Now().UTC().Format(time.RFC3339)
	os.WriteFile(file, []byte(fixture(100, at, "ue4ss_live")), 0600)
	s := Server{LogPath: file}
	d, e := s.Data()
	if e != nil || len(d.Points) != 1 || d.Points[0].Position[0] != 100 || d.Status.Stale {
		t.Fatal(d, e)
	}
	os.WriteFile(file, []byte(fixture(200, at, "ue4ss_live")), 0600)
	d, e = s.Data()
	if e != nil || d.Points[0].Position[0] != 200 {
		t.Fatal("old position cached", e)
	}
	old := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	os.WriteFile(file, []byte(fixture(300, old, "ue4ss_live")+"irrelevant current log line\n"), 0600)
	d, e = s.Data()
	if e != nil || !d.Status.Stale || !d.Points[0].Stale {
		t.Fatal("logmtime faked freshness", d, e)
	}
	os.WriteFile(file, []byte(fixture(100, at, "ue4ss_discovery")), 0600)
	d, e = s.Data()
	if e != nil || len(d.Points) != 0 || !d.Status.Stale {
		t.Fatal("bounded scan used as live", d, e)
	}
}
func TestAPIOnlyConfiguredInputAndLoopback(t *testing.T) {
	file := filepath.Join(t.TempDir(), "UE4SS.log")
	os.WriteFile(file, []byte(fixture(100, time.Now().UTC().Format(time.RFC3339), "ue4ss_live")), 0600)
	h := (Server{LogPath: file}).Handler()
	request := func(method, url, host string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(method, url, nil)
		r.Host = host
		h.ServeHTTP(w, r)
		return w
	}
	w := request("GET", "/api/map?path=secret", "127.0.0.1:18126")
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	var d Data
	if e := json.Unmarshal(w.Body.Bytes(), &d); e != nil || len(d.Points) != 1 {
		t.Fatal(e)
	}
	if request("GET", "/UE4SS.log", "localhost").Code != 404 || request("POST", "/api/map", "localhost").Code != 405 || request("GET", "/api/map", "attacker.example").Code != 403 {
		t.Fatal("unsafe route")
	}
	os.WriteFile(file, []byte(strings.Repeat("x", 64*1024+1)+"\n"), 0600)
	if _, e := (Server{LogPath: file}).Data(); e == nil {
		t.Fatal("oversize line accepted")
	}
}

func TestSessionWaitingAndBoundedTail(t *testing.T) {
	p := filepath.Join(t.TempDir(), "UE4SS.log")
	at := time.Now().UTC().Format(time.RFC3339)
	heartbeat := fmt.Sprintf(`BWADVISOR {"protocol_version":"0.3","source":"ue4ss_live","event":"heartbeat","data":{"scan_id":"synthetic-session","session_id":"synthetic-session","captured_at":"%s","sequence":500,"waiting_for_player":true}}`+"\n", at)
	os.WriteFile(p, []byte(fixture(100, at, "ue4ss_live")+heartbeat), 0600)
	d, e := (Server{LogPath: p}).Data()
	if e != nil || len(d.Points) != 0 || !d.Status.Waiting {
		t.Fatal("waiting did not clear", d, e)
	}
	old := fixture(100, at, "ue4ss_live")
	newSession := strings.ReplaceAll(fixture(200, at, "ue4ss_live"), "synthetic-session", "new-session")
	os.WriteFile(p, []byte(old+newSession), 0600)
	d, e = (Server{LogPath: p}).Data()
	if e != nil || len(d.Points) != 1 || d.Status.SessionID != "new-session" || d.Points[0].Position[0] != 200 {
		t.Fatal("old session retained", d, e)
	}
	// Repeated short unrelated lines keep Scanner bounded while forcing tail mode.
	large := strings.Repeat("unrelated\n", int(MaxLogBytes)/10+100) + newSession + "BWADVISOR {partial"
	os.WriteFile(p, []byte(large), 0600)
	d, e = (Server{LogPath: p}).Data()
	if e != nil || d.Status.SkippedOlderBytes <= 0 || len(d.Points) != 1 {
		t.Fatal("tail/partial handling", d, e)
	}
}

func TestFriendlyRuntimeLabels(t *testing.T) {
	if got := actorLabel("MiningCamp_C", "World.Level.MiningCamp_C_214"); got != "Mining camp · MiningCamp_C_214" {
		t.Fatal(got)
	}
	if actorLabel("player", "runtimePlayer") != "Player" {
		t.Fatal("player label")
	}
}
