package main

import (
	"encoding/json"
	"testing"
)

func TestResourceObservedCoordinates(t *testing.T) {
	for _, raw := range []string{`{}`, `{"position":null}`, `{"position":[null,0,0]}`, `{"position":[0,0]}`, `{"position":[0,0,1e9]}`} {
		var item resourceItem
		if json.Unmarshal([]byte(raw), &item) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	var item resourceItem
	if e := json.Unmarshal([]byte(`{"id":"a","position":[0,0,0]}`), &item); e != nil || len(item.Position) != 3 {
		t.Fatal(e)
	}
}
func TestPlacementObservedMetrics(t *testing.T) {
	for _, raw := range []string{`{}`, `{"position":[0,0,0],"distance_to_core_engine_units":0}`, `{"position":[0,0,0],"baseline_food_node_count":null,"distance_to_core_engine_units":0}`, `{"position":[0,0,0],"baseline_food_node_count":0}`, `{"position":[0,0,0],"baseline_food_node_count":0,"distance_to_core_engine_units":null}`, `{"position":[0,0,0],"baseline_food_node_count":0,"distance_to_core_engine_units":-1}`, `{"position":[0,0,0],"baseline_food_node_count":-1,"distance_to_core_engine_units":0}`, `{"position":[null,0,0],"baseline_food_node_count":0,"distance_to_core_engine_units":0}`} {
		var item placementSite
		if json.Unmarshal([]byte(raw), &item) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	var item placementSite
	if e := json.Unmarshal([]byte(`{"position":[0,0,0],"baseline_food_node_count":0,"distance_to_core_engine_units":0}`), &item); e != nil {
		t.Fatal(e)
	}
}

func TestGearCraftRequiresObservedQuantities(t *testing.T) {
	for _, raw := range []string{`{"item":"Iron_C","deficit":1}`, `{"item":"Iron_C","need":1,"equipped":0,"stored":null,"reserve":0,"target":1,"have":0,"deficit":1}`} {
		var item gearCraft
		if json.Unmarshal([]byte(raw), &item) == nil {
			t.Fatalf("accepted unknown quantities: %s", raw)
		}
	}
	var item gearCraft
	if err := json.Unmarshal([]byte(`{"item":"Iron_C","need":1,"equipped":0,"stored":0,"reserve":0,"target":1,"have":0,"deficit":1}`), &item); err != nil {
		t.Fatal(err)
	}
}
