package livemap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestResourceEvidenceAndBounds(t *testing.T) {
	p := filepath.Join(t.TempDir(), "resource.json")
	b := ResourceBaseline{Schema: "bellwright.harvestable-baseline.v1", Source: "packaged_game_asset", Unknowns: []string{"availability unknown"}, Nodes: []ResourceNode{{ID: "synthetic", Class: "SyntheticHarvestable", Position: [3]float64{1, 2, 3}}}}
	write := func() { raw, _ := json.Marshal(b); os.WriteFile(p, raw, 0600) }
	write()
	s := Server{ResourcePath: p}
	got, e := s.Resources()
	if e != nil || len(got.Nodes) != 1 {
		t.Fatal(got, e)
	}
	b.Source = "ue4ss_live"
	write()
	if _, e = s.Resources(); e == nil {
		t.Fatal("packaged nodes accepted as live")
	}
	b.Source = "packaged_game_asset"
	b.Nodes[0].Position[0] = 1e10
	write()
	if _, e = s.Resources(); e == nil {
		t.Fatal("unbounded coordinate accepted")
	}
	b.Nodes[0].Position[0] = 1
	b.Nodes[0].Class = ""
	write()
	if _, e = s.Resources(); e == nil {
		t.Fatal("missing asset provenance accepted")
	}
}
func TestIncompleteResourceCoordinatesRejected(t *testing.T) {
	for _, raw := range []string{`{"id":"x","class_asset_path":"Fixture"}`, `{"id":"x","class_asset_path":"Fixture","position":null}`, `{"id":"x","class_asset_path":"Fixture","position":[1,2]}`} {
		var n ResourceNode
		if json.Unmarshal([]byte(raw), &n) == nil {
			t.Fatal("partial evidence accepted", raw)
		}
	}
}
func TestPlacementEvidence(t *testing.T) {
	p := filepath.Join(t.TempDir(), "placements.json")
	site := PlacementSite{Position: [3]float64{1, 2, 3}, Count: 10}
	ps := []Placement{{Schema: "bellwright.foraging-placement-candidate.v1", ID: "SyntheticCamp", Radius: 15000, Current: site, Practical: site, Assumptions: []string{"units and availability unverified"}}}
	raw, _ := json.Marshal(ps)
	os.WriteFile(p, raw, 0600)
	s := Server{PlacementPath: p}
	got, e := s.Placements()
	if e != nil || len(got) != 1 {
		t.Fatal(got, e)
	}
	ps[0].Assumptions = nil
	raw, _ = json.Marshal(ps)
	os.WriteFile(p, raw, 0600)
	if _, e = s.Placements(); e == nil {
		t.Fatal("assumptions omitted")
	}
}
