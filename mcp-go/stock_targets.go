package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"math"
	"sort"
	"strings"
)

type stockTargetQuery struct {
	Targets map[string]*float64 `json:"targets"`
	Scope   string              `json:"scope"`
	Limit   *int                `json:"limit,omitempty"`
}
type stockTargetItem struct {
	Item     string   `json:"item"`
	Target   float64  `json:"target"`
	Quantity *float64 `json:"quantity"`
	Deficit  *float64 `json:"deficit"`
	Surplus  *float64 `json:"surplus"`
	State    string   `json:"state"`
}
type stockTargetOutput struct {
	Source       string            `json:"source"`
	SnapshotID   *int64            `json:"snapshot_id"`
	IngestedAt   *string           `json:"ingested_at"`
	RetrievedAt  string            `json:"retrieved_at"`
	Scope        string            `json:"scope"`
	State        string            `json:"state"`
	Total        int               `json:"total"`
	Returned     int               `json:"returned"`
	Truncated    int               `json:"truncated"`
	Limit        int               `json:"limit"`
	Unknown      int               `json:"unknown"`
	DeficitItems int               `json:"deficit_items"`
	Items        []stockTargetItem `json:"items"`
	Note         string            `json:"note"`
}

func registerStockTargetTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "compare_saved_stock_targets", Description: "Compare 1 through 100 supplied exact item targets against one explicit saved scope: stored, npc_carried or equipped. Targets are nonnegative integer counts. No persistent settings or summed scopes. Limit defaults 25, maximum 100.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.stockTargets)
}
func (b *backend) stockTargets(ctx context.Context, _ *mcp.CallToolRequest, q stockTargetQuery) (*mcp.CallToolResult, stockTargetOutput, error) {
	o := stockTargetOutput{Source: "companion_save", Scope: q.Scope, State: "unknown", Limit: 25, Items: []stockTargetItem{}, Note: "Latest ingested save only; ingestion/retrieval are not capture time. One selected scope, never summed: stored totals cover world containers; npc_carried is the backend aggregate; equipped counts explicit recruited NPC equipment slots. For an observed complete aggregate, absent item keys are assumed zero. Null/malformed quantities remain unknown. Equipped completeness requires observed ownership and equipment for every recruited NPC. No recipes, yields or persisted targets."}
	if q.Limit != nil {
		o.Limit = *q.Limit
	}
	if o.Limit < 1 || o.Limit > 100 || len(q.Targets) < 1 || len(q.Targets) > 100 {
		return nil, o, fmt.Errorf("require 1 through 100 targets and limit 1 through 100")
	}
	if q.Scope != "stored" && q.Scope != "npc_carried" && q.Scope != "equipped" {
		return nil, o, fmt.Errorf("scope must be stored, npc_carried or equipped")
	}
	keys := []string{}
	for k, v := range q.Targets {
		if strings.TrimSpace(k) == "" || len(k) > 256 || v == nil || math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 || math.Trunc(*v) != *v || *v > 9007199254740991 {
			return nil, o, fmt.Errorf("targets require exact nonempty item keys and finite nonnegative safe integer counts")
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	o.Total = len(keys)
	_, raw, e := b.read(ctx, "/api/world", "companion_save", "")
	if e != nil {
		return nil, o, fmt.Errorf("saved world unavailable")
	}
	var w struct {
		SnapshotID *int64  `json:"snapshot_id"`
		IngestedAt *string `json:"ingested_at"`
		Storage    *struct {
			Totals  map[string]any `json:"totals"`
			Carried map[string]any `json:"npc_carried"`
		} `json:"storage"`
		NPCs *[]map[string]any `json:"npcs"`
	}
	if decodeEvidence(raw.Data, &w) != nil {
		return nil, o, fmt.Errorf("invalid saved inventory evidence")
	}
	o.SnapshotID = w.SnapshotID
	o.IngestedAt = w.IngestedAt
	o.RetrievedAt = raw.RetrievedAt
	var amounts map[string]any
	complete := false
	if q.Scope == "equipped" {
		if w.NPCs != nil {
			amounts = map[string]any{}
			complete = true
			for _, n := range *w.NPCs {
				player, ok := n["is_player_npc"].(bool)
				if !ok {
					complete = false
					continue
				}
				if !player {
					continue
				}
				eq, ok := n["equipment"].(map[string]any)
				if !ok {
					complete = false
					continue
				}
				for _, v := range eq {
					item, ok := v.(string)
					if !ok {
						complete = false
						continue
					}
					if item != "" {
						count, _ := amounts[item].(float64)
						amounts[item] = count + 1
					}
				}
			}
		}
	} else if w.Storage != nil {
		if q.Scope == "stored" {
			amounts = w.Storage.Totals
		} else {
			amounts = w.Storage.Carried
		}
		complete = amounts != nil
	}
	// Invalid observed values defeat the complete-total zero assumption but never erase known quantities.
	valid := func(v any) (float64, bool) {
		n, ok := v.(float64)
		return n, ok && !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0 && math.Trunc(n) == n && n <= 9007199254740991
	}
	for _, v := range amounts {
		if _, ok := valid(v); !ok {
			complete = false
		}
	}
	if complete {
		o.State = "available"
	}
	for _, k := range keys {
		row := stockTargetItem{Item: k, Target: *q.Targets[k], State: "unknown"}
		v, present := amounts[k]
		n, ok := valid(v)
		if q.Scope == "equipped" && !complete {
			ok = false
		}
		if !present && complete {
			n = 0
			ok = true
		}
		if ok {
			row.Quantity = &n
			d, s := math.Max(row.Target-n, 0), math.Max(n-row.Target, 0)
			row.Deficit = &d
			row.Surplus = &s
			row.State = "met"
			if d > 0 {
				row.State = "deficit"
				o.DeficitItems++
			}
		} else {
			o.Unknown++
		}
		if len(o.Items) < o.Limit {
			o.Items = append(o.Items, row)
		}
	}
	o.Returned = len(o.Items)
	o.Truncated = o.Total - o.Returned
	return nil, o, nil
}
