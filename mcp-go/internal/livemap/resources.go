package livemap

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

type evidencePosition [3]float64

func (p *evidencePosition) UnmarshalJSON(raw []byte) error {
	var values []float64
	if e := json.Unmarshal(raw, &values); e != nil {
		return e
	}
	if len(values) != 3 {
		return fmt.Errorf("position must have three observed coordinates")
	}
	copy(p[:], values)
	return nil
}

type ResourceNode struct {
	ID       string           `json:"id"`
	Category string           `json:"baseline_category"`
	Class    string           `json:"class_asset_path"`
	Position evidencePosition `json:"position"`
}

func (n *ResourceNode) UnmarshalJSON(raw []byte) error {
	type plain ResourceNode
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return e
	}
	if _, ok := fields["position"]; !ok {
		return fmt.Errorf("missing resource position")
	}
	return json.Unmarshal(raw, (*plain)(n))
}

type ResourceBaseline struct {
	Schema   string         `json:"schema"`
	Source   string         `json:"source_kind"`
	Nodes    []ResourceNode `json:"nodes"`
	Unknowns []string       `json:"unknowns"`
}
type PlacementSite struct {
	Position evidencePosition `json:"position"`
	Count    int              `json:"baseline_food_node_count"`
	Distance float64          `json:"distance_to_core_engine_units"`
}

func (p *PlacementSite) UnmarshalJSON(raw []byte) error {
	type plain PlacementSite
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(raw, &fields); e != nil {
		return e
	}
	if _, ok := fields["position"]; !ok {
		return fmt.Errorf("missing candidate position")
	}
	return json.Unmarshal(raw, (*plain)(p))
}

type Placement struct {
	Schema      string        `json:"schema"`
	ID          string        `json:"target_runtime_actor_id"`
	Radius      float64       `json:"radius_engine_units_assumed"`
	Current     PlacementSite `json:"current"`
	Practical   PlacementSite `json:"practical_candidate"`
	Assumptions []string      `json:"assumptions"`
}

func (s Server) Placements() ([]Placement, error) {
	f, e := os.Open(s.PlacementPath)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, 1048577))
	if e != nil || len(raw) > 1048576 {
		return nil, fmt.Errorf("oversized placements")
	}
	var ps []Placement
	if e = json.Unmarshal(raw, &ps); e != nil {
		return nil, e
	}
	if len(ps) > 256 {
		return nil, fmt.Errorf("too many placements")
	}
	for _, p := range ps {
		if p.Schema != "bellwright.foraging-placement-candidate.v1" || p.ID == "" || len(p.Assumptions) == 0 || p.Radius <= 0 || p.Radius > 1e6 {
			return nil, fmt.Errorf("invalid placement evidence")
		}
		for _, site := range []PlacementSite{p.Current, p.Practical} {
			if site.Count < 0 {
				return nil, fmt.Errorf("invalid count")
			}
			for _, v := range site.Position {
				if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e8 {
					return nil, fmt.Errorf("invalid coordinate")
				}
			}
		}
	}
	return ps, nil
}

func (s Server) Resources() (ResourceBaseline, error) {
	var b ResourceBaseline
	f, e := os.Open(s.ResourcePath)
	if e != nil {
		return b, e
	}
	defer f.Close()
	raw, e := io.ReadAll(io.LimitReader(f, 50*1024*1024+1))
	if e != nil || len(raw) > 50*1024*1024 {
		return b, fmt.Errorf("oversized baseline")
	}
	if e = json.Unmarshal(raw, &b); e != nil {
		return b, e
	}
	if b.Schema != "bellwright.harvestable-baseline.v1" || b.Source != "packaged_game_asset" || len(b.Nodes) > 100000 || len(b.Unknowns) == 0 {
		return b, fmt.Errorf("invalid baseline provenance or bound")
	}
	for _, n := range b.Nodes {
		if n.ID == "" || n.Class == "" {
			return b, fmt.Errorf("missing resource evidence")
		}
		for _, v := range n.Position {
			if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e8 {
				return b, fmt.Errorf("invalid coordinate")
			}
		}
	}
	return b, nil
}
