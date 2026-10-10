package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
	"regexp"
	"sort"
	"strings"
)

type observedItemQuery struct {
	Query string `json:"query,omitempty" jsonschema:"Case insensitive substring of saved class or formatted identifier"`
	Limit *int   `json:"limit,omitempty"`
}
type observedItemScope struct {
	Source   string  `json:"source"`
	Quantity float64 `json:"quantity"`
}
type observedItem struct {
	Class         string              `json:"item_class"`
	Display       string              `json:"display_identifier"`
	OfficialLabel *string             `json:"official_label"`
	LabelSource   string              `json:"label_source"`
	Sources       []observedItemScope `json:"sources"`
}
type observedItemsOutput struct {
	Source        string         `json:"source"`
	SnapshotID    *int           `json:"snapshot_id"`
	IngestedAt    *string        `json:"ingested_at"`
	RetrievedAt   string         `json:"retrieved_at"`
	Total         int            `json:"total"`
	Matched       int            `json:"matched"`
	Returned      int            `json:"returned"`
	Truncated     int            `json:"truncated"`
	Limit         int            `json:"limit"`
	Items         []observedItem `json:"items"`
	UnknownScopes []string       `json:"unknown_scopes"`
	Note          string         `json:"note"`
}

var itemNumberSuffix = regexp.MustCompile(`_\d+$`)
var itemLowerUpper = regexp.MustCompile(`([a-z])([A-Z])`)
var itemAcronym = regexp.MustCompile(`([A-Z])([A-Z][a-z])`)

// Matches the local UI's identifier formatting; it does not resolve game names.
func observedDisplayIdentifier(id string) string {
	id = strings.TrimSuffix(id, "_C")
	id = itemNumberSuffix.ReplaceAllString(id, "")
	id = strings.ReplaceAll(id, "_", " ")
	id = itemLowerUpper.ReplaceAllString(id, "${1} ${2}")
	id = itemAcronym.ReplaceAllString(id, "${1} ${2}")
	return strings.TrimSpace(id)
}
func registerObservedItemTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "search_observed_items", Description: "Search positive-quantity saved item classes and occupied equipment slots. Case-insensitive substring matches raw class or formatted identifier; formatted identifiers are not verified game names. Bounded default25/max100; source scopes overlap and must not be summed. Preset-only references excluded; no full game catalog claim.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.observedItems)
}

type observedCarried struct {
	Item *string  `json:"item"`
	Qty  *float64 `json:"qty"`
}

func (b *backend) observedItems(ctx context.Context, _ *mcp.CallToolRequest, q observedItemQuery) (*mcp.CallToolResult, observedItemsOutput, error) {
	var out observedItemsOutput
	limit, err := (spatialQuery{Limit: q.Limit}).validate()
	if err != nil {
		return nil, out, err
	}
	if len(q.Query) > 256 {
		return nil, out, fmt.Errorf("query exceeds 256 bytes")
	}
	_, raw, err := b.read(ctx, "/api/world", "companion_save", "")
	if err != nil {
		return nil, out, err
	}
	var w struct {
		SnapshotID *int    `json:"snapshot_id"`
		IngestedAt *string `json:"ingested_at"`
		Storage    *struct {
			Totals     map[string]*float64 `json:"totals"`
			NPCCarried map[string]*float64 `json:"npc_carried"`
			Containers []struct {
				Items map[string]*float64 `json:"items"`
			} `json:"containers"`
		} `json:"storage"`
		NPCs []struct {
			Equipment map[string]*string `json:"equipment"`
		} `json:"npcs"`
		Player *struct {
			Equipment map[string]*string `json:"equipment"`
			Carried   []observedCarried  `json:"carried"`
		} `json:"player"`
		Carried map[string][]observedCarried `json:"carried"`
	}
	if err = decodeEvidence(raw.Data, &w); err != nil {
		return nil, out, err
	}
	if w.Storage == nil && w.NPCs == nil && w.Player == nil && w.Carried == nil {
		return nil, out, fmt.Errorf("missing saved item evidence")
	}
	if len(w.NPCs) > 100000 || len(w.Carried) > 100000 {
		return nil, out, fmt.Errorf("saved item evidence exceeds bound")
	}
	catalog := map[string]map[string]float64{}
	unknown := map[string]bool{}
	add := func(id, scope string, qty float64) error {
		if id == "" || len(id) > 1024 || qty < 0 || math.IsNaN(qty) || math.IsInf(qty, 0) {
			return fmt.Errorf("invalid saved item identity or quantity")
		}
		if qty == 0 {
			return nil
		}
		if catalog[id] == nil {
			catalog[id] = map[string]float64{}
		}
		v := catalog[id][scope] + qty
		if math.IsInf(v, 0) {
			return fmt.Errorf("saved quantity overflow")
		}
		catalog[id][scope] = v
		return nil
	}
	quantities := func(m map[string]*float64, scope string) error {
		if m == nil {
			unknown[scope] = true
			return nil
		}
		if len(m) > 100000 {
			return fmt.Errorf("too many saved item entries")
		}
		for id, v := range m {
			if v == nil {
				return fmt.Errorf("missing saved quantity")
			}
			if err := add(id, scope, *v); err != nil {
				return err
			}
		}
		return nil
	}
	equipment := func(m map[string]*string, scope string) error {
		if m == nil {
			unknown[scope] = true
			return nil
		}
		for _, id := range m {
			if id == nil {
				return fmt.Errorf("null saved equipment identity")
			}
			if *id != "" {
				if err := add(*id, scope, 1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	carried := func(items []observedCarried, scope string) error {
		if items == nil {
			unknown[scope] = true
			return nil
		}
		if len(items) > 100000 {
			return fmt.Errorf("too many carried entries")
		}
		for _, i := range items {
			if i.Item == nil || i.Qty == nil {
				return fmt.Errorf("missing carried identity or quantity")
			}
			if err := add(*i.Item, scope, *i.Qty); err != nil {
				return err
			}
		}
		return nil
	}
	if w.Storage == nil {
		for _, s := range []string{"storage_totals", "container_items", "npc_carried_totals"} {
			unknown[s] = true
		}
	} else {
		if err = quantities(w.Storage.Totals, "storage_totals"); err != nil {
			return nil, out, err
		}
		if err = quantities(w.Storage.NPCCarried, "npc_carried_totals"); err != nil {
			return nil, out, err
		}
		if w.Storage.Containers == nil {
			unknown["container_items"] = true
		}
		if len(w.Storage.Containers) > 100000 {
			return nil, out, fmt.Errorf("too many containers")
		}
		for _, c := range w.Storage.Containers {
			if err = quantities(c.Items, "container_items"); err != nil {
				return nil, out, err
			}
		}
	}
	if w.NPCs == nil {
		unknown["saved_npc_equipment"] = true
	}
	for _, n := range w.NPCs {
		if err = equipment(n.Equipment, "saved_npc_equipment"); err != nil {
			return nil, out, err
		}
	}
	if w.Player == nil {
		unknown["player_equipment"] = true
		unknown["player_carried"] = true
	} else {
		if err = equipment(w.Player.Equipment, "player_equipment"); err != nil {
			return nil, out, err
		}
		if err = carried(w.Player.Carried, "player_carried"); err != nil {
			return nil, out, err
		}
	}
	if w.Carried == nil {
		unknown["actor_carried"] = true
	}
	for _, items := range w.Carried {
		if err = carried(items, "actor_carried"); err != nil {
			return nil, out, err
		}
	}
	out = observedItemsOutput{Source: raw.Source, SnapshotID: w.SnapshotID, IngestedAt: w.IngestedAt, RetrievedAt: raw.RetrievedAt, Total: len(catalog), Limit: limit, Items: []observedItem{}, UnknownScopes: []string{}, Note: "Latest ingested save only, not live stock or a full game catalog. Positive quantities and occupied saved equipment slots only; preset-only references excluded. Source scopes overlap (including storage totals/containers and actor/player/NPC carried), so do not sum them. Equipment quantities count occupied slots across saved NPCs, not recruited ownership. Missing scopes remain unknown; listed scope quantities may be partial when that scope is unknown. Display identifiers only format class strings; official game labels are unknown. Retrieval time is not save time."}
	for s := range unknown {
		out.UnknownScopes = append(out.UnknownScopes, s)
	}
	sort.Strings(out.UnknownScopes)
	ids := make([]string, 0, len(catalog))
	for id := range catalog {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	query := strings.ToLower(q.Query)
	for _, id := range ids {
		display := observedDisplayIdentifier(id)
		if !strings.Contains(strings.ToLower(id), query) && !strings.Contains(strings.ToLower(display), query) {
			continue
		}
		out.Matched++
		if len(out.Items) >= limit {
			continue
		}
		sources := []observedItemScope{}
		keys := []string{}
		for s := range catalog[id] {
			keys = append(keys, s)
		}
		sort.Strings(keys)
		for _, s := range keys {
			sources = append(sources, observedItemScope{s, catalog[id][s]})
		}
		out.Items = append(out.Items, observedItem{Class: id, Display: display, LabelSource: "identifier_format", Sources: sources})
	}
	out.Returned = len(out.Items)
	out.Truncated = out.Matched - out.Returned
	return nil, out, nil
}
