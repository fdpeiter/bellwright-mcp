package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
)

func (g *gearCraft) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"item", "need", "equipped", "stored", "reserve", "target", "have", "deficit"} {
		value, present := fields[key]
		if !present || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("gear craft requires observed %s", key)
		}
	}
	type plain gearCraft
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*g = gearCraft(value)
	return nil
}

// Pointer-backed decoding keeps missing/null observations from becoming zero.
func evidencePosition(p []*float64) ([]float64, error) {
	if len(p) != 3 {
		return nil, fmt.Errorf("position requires three observed coordinates")
	}
	out := make([]float64, 3)
	for i, v := range p {
		if v == nil || !validCoordinate(*v) {
			return nil, fmt.Errorf("position requires finite observed coordinates")
		}
		out[i] = *v
	}
	return out, nil
}
func (r *resourceItem) UnmarshalJSON(data []byte) error {
	var v struct {
		ID       string     `json:"id"`
		Category string     `json:"baseline_category"`
		Class    string     `json:"class_asset_path"`
		Position []*float64 `json:"position"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	p, err := evidencePosition(v.Position)
	if err != nil {
		return err
	}
	*r = resourceItem{ID: v.ID, Category: v.Category, Class: v.Class, Position: p}
	return nil
}
func (p *placementSite) UnmarshalJSON(data []byte) error {
	var v struct {
		Position []*float64 `json:"position"`
		Count    *int       `json:"baseline_food_node_count"`
		Distance *float64   `json:"distance_to_core_engine_units"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	position, err := evidencePosition(v.Position)
	if err != nil {
		return err
	}
	if v.Count == nil || *v.Count < 0 {
		return fmt.Errorf("placement requires observed nonnegative node count")
	}
	if v.Distance == nil || math.IsNaN(*v.Distance) || math.IsInf(*v.Distance, 0) || *v.Distance < 0 {
		return fmt.Errorf("placement requires finite nonnegative observed distance")
	}
	*p = placementSite{Position: position, Count: *v.Count, Distance: *v.Distance}
	return nil
}
