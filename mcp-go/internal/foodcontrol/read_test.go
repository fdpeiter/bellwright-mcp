package foodcontrol

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadFreshStaleAndMenu(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.log")
	now := time.Date(2026, 10, 11, 1, 0, 0, 0, time.UTC)
	line := `BWFOOD {"source":"ue4ss_live","session_id":"one","captured_at":"2026-10-11T01:00:00Z","available":true,"status":"observed","write_available":true,"buildings":[],"settlements":[]}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0600); err != nil {
		t.Fatal(err)
	}
	got, e := Read(path, now)
	if e != nil || !got.Available || got.Stale || got.WriteAvailable {
		t.Fatalf("fresh: %#v %v", got, e)
	}
	got, e = Read(path, now.Add(21*time.Second))
	if e != nil || got.Available || !got.Stale || got.Status != "stale" {
		t.Fatalf("stale: %#v %v", got, e)
	}
	menu := `BWFOOD {"source":"ue4ss_live","session_id":"two","captured_at":"2026-10-11T01:00:02Z","available":false,"status":"waiting_for_player"}` + "\n"
	if e = os.WriteFile(path, []byte(line+menu), 0600); e != nil {
		t.Fatal(e)
	}
	got, e = Read(path, now.Add(3*time.Second))
	if e != nil || got.Available || got.Data.SessionID != "two" || len(got.Data.Buildings) > 0 {
		t.Fatalf("menu: %#v %v", got, e)
	}
}
func TestRejectInvalidAndFuture(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.log")
	raw := "BWFOOD invalid\n" + `BWFOOD {"source":"companion_save","session_id":"a","captured_at":"2026-10-11T01:00:00Z","available":true}` + "\n" + `BWFOOD {"source":"ue4ss_live","session_id":"a","captured_at":"2026-10-11T02:00:00Z","available":true}` + "\n"
	if e := os.WriteFile(path, []byte(raw), 0600); e != nil {
		t.Fatal(e)
	}
	got, e := Read(path, time.Date(2026, 10, 11, 1, 0, 0, 0, time.UTC))
	if e != nil || got.Data != nil || got.Available || !got.Stale {
		t.Fatalf("invalid: %#v %v", got, e)
	}
}
