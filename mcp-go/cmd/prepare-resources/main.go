// Convert owner-exported packaged foliage to the live-map's static baseline.
// No saves/runtime reads. Positions remain unverified packaged translations.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

type node struct {
	ID       string     `json:"id"`
	Category string     `json:"baseline_category"`
	Class    string     `json:"class_asset_path"`
	Position [3]float64 `json:"position"`
}
type baseline struct {
	Schema    string   `json:"schema"`
	Source    string   `json:"source_kind"`
	InputHash string   `json:"input_sha256"`
	Nodes     []node   `json:"nodes"`
	Unknowns  []string `json:"unknowns"`
}
type exported struct {
	Type       string
	Name       string
	Properties struct {
		Class struct {
			Path string `json:"ObjectPath"`
		} `json:"FoliageComponentClass"`
		Instances []struct {
			Transform struct{ Translation struct{ X, Y, Z *float64 } }
		}
	}
}

func convert(r io.Reader) (baseline, error) {
	out := baseline{Schema: "bellwright.harvestable-baseline.v1", Source: "packaged_game_asset", Nodes: []node{}, Unknowns: []string{
		"Packaged baseline only; availability, depletion, regrowth and runtime changes unknown.",
		"Instance translation interpreted as world coordinates; parent transforms and ground offsets are not applied or verified.",
		"Class folder supplies category evidence, not item yields, ownership, reachability or measured production.",
		"Input export and selected loaded packages do not establish complete world coverage.",
	}}
	hash := sha256.New()
	decoder := json.NewDecoder(io.TeeReader(io.LimitReader(r, 1024*1024*1024+1), hash))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('[') {
		return out, fmt.Errorf("expected package export array")
	}
	seen := map[string]bool{}
	for decoder.More() {
		var item exported
		if err = decoder.Decode(&item); err != nil {
			return out, fmt.Errorf("invalid export")
		}
		if decoder.InputOffset() > 1024*1024*1024 {
			return out, fmt.Errorf("input exceeds 1 GiB")
		}
		path := item.Properties.Class.Path
		if item.Type != "MistFoliageDataComponent" || !strings.Contains(path, "/FoliageHarvestable/") {
			continue
		}
		if item.Name == "" {
			return out, fmt.Errorf("component identity missing")
		}
		category := "harvestable_unmapped"
		if at := strings.Index(path, "/Harvestable/"); at >= 0 {
			category = strings.Split(path[at+len("/Harvestable/"):], "/")[0]
			if category == "" {
				return out, fmt.Errorf("category missing")
			}
		}
		for index, instance := range item.Properties.Instances {
			p := instance.Transform.Translation
			if p.X == nil || p.Y == nil || p.Z == nil {
				return out, fmt.Errorf("observed translation missing")
			}
			position := [3]float64{*p.X, *p.Y, *p.Z}
			for _, v := range position {
				if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e8 {
					return out, fmt.Errorf("invalid translation")
				}
			}
			id := fmt.Sprintf("%s:%d", item.Name, index)
			if seen[id] {
				return out, fmt.Errorf("duplicate component identity")
			}
			if len(out.Nodes) >= 100000 {
				return out, fmt.Errorf("baseline exceeds 100000 nodes")
			}
			seen[id] = true
			out.Nodes = append(out.Nodes, node{id, category, path, position})
		}
	}
	if _, err = decoder.Token(); err != nil {
		return out, fmt.Errorf("incomplete export array")
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return out, fmt.Errorf("trailing data or oversized input")
	}
	if decoder.InputOffset() > 1024*1024*1024 {
		return out, fmt.Errorf("input exceeds 1 GiB")
	}
	if len(out.Nodes) == 0 {
		return out, fmt.Errorf("no harvestable instances; choose the correct map package")
	}
	out.InputHash = hex.EncodeToString(hash.Sum(nil))
	return out, nil
}

func run(input, output string) error {
	f, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("cannot read input")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1024*1024*1024 {
		return fmt.Errorf("invalid bounded export file")
	}
	result, err := convert(f)
	if err != nil {
		return err
	}
	target, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("output unavailable or already exists")
	}
	err = json.NewEncoder(target).Encode(result)
	closeErr := target.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: prepare-resources MAP_EXPORT.json PRIVATE_BASELINE.json")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Prepared packaged baseline; resource availability and world coverage remain unknown.")
}
