package main

import (
	"strings"
	"testing"
)

const fixture = `{"Type":"MistFoliageDataComponent","Name":"synthetic","Properties":{"FoliageComponentClass":{"ObjectPath":"/Game/FoliageHarvestable/Harvestable/Berry/Bush"},"Instances":[{"Transform":{"Translation":{"X":1,"Y":2,"Z":3}}}]}}`

func TestPackagedBaselineKeepsProvenanceAndCoordinates(t *testing.T) {
	result, err := convert(strings.NewReader("[" + fixture + "]"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Source != "packaged_game_asset" || len(result.Nodes) != 1 || result.Nodes[0].Category != "Berry" || result.Nodes[0].Position != [3]float64{1, 2, 3} || len(result.InputHash) != 64 || len(result.Unknowns) == 0 {
		t.Fatalf("unexpected synthetic baseline: %+v", result)
	}
}
func TestInvalidOrAmbiguousEvidenceFails(t *testing.T) {
	for _, input := range []string{
		"[" + fixture + "," + fixture + "]",
		"[" + strings.Replace(fixture, `"Z":3`, `"other":3`, 1) + "]",
		"[" + strings.Replace(fixture, `"X":1`, `"X":100000001`, 1) + "]",
		"[" + strings.Replace(fixture, "FoliageHarvestable", "FoliageAttackable", 1) + "]",
		"[" + fixture + "]{}", "[]", "[" + fixture,
	} {
		if _, err := convert(strings.NewReader(input)); err == nil {
			t.Fatal("accepted invalid evidence")
		}
	}
}
