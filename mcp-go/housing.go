package main

import (
	"context"
	"fmt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type housingQuery struct {
	Limit *int `json:"limit,omitempty"`
}
type housingType struct {
	Class    string `json:"cls"`
	Label    string `json:"label"`
	Capacity *int   `json:"cap"`
	Count    *int   `json:"count"`
	Beds     *int   `json:"beds"`
}
type housingPerson struct {
	GUID   *string `json:"guid"`
	Housed *bool   `json:"housed"`
}
type housingOutput struct {
	Source              string          `json:"source"`
	SnapshotID          *int64          `json:"snapshot_id"`
	IngestedAt          *string         `json:"ingested_at"`
	RetrievedAt         string          `json:"retrieved_at"`
	State               string          `json:"state"`
	Quarters            *int            `json:"quarters"`
	Houses              *int            `json:"houses"`
	Recruited           *int            `json:"recruited"`
	CapacityDeficit     *int            `json:"capacity_deficit"`
	SpareCapacity       *int            `json:"spare_capacity"`
	UnknownOwnership    int             `json:"unknown_ownership"`
	ObservedHoused      int             `json:"observed_housed"`
	ObservedNotHoused   int             `json:"observed_not_housed"`
	UnknownHousing      int             `json:"unknown_housing"`
	TypeTotal           int             `json:"type_total"`
	TypeReturned        int             `json:"type_returned"`
	TypeTruncated       int             `json:"type_truncated"`
	Types               []housingType   `json:"types"`
	IndividualTotal     int             `json:"individual_total"`
	IndividualReturned  int             `json:"individual_returned"`
	IndividualTruncated int             `json:"individual_truncated"`
	Individuals         []housingPerson `json:"individuals"`
	Limit               int             `json:"limit"`
	Note                string          `json:"note"`
}

func registerHousingTools(s *mcp.Server, b *backend) {
	mcp.AddTool(s, &mcp.Tool{Name: "read_saved_housing", Description: "Read saved aggregate bed capacity and separate individual housing observations for explicitly recruited NPCs. Each list defaults to 25, maximum 100. Missing housing is unknown.", Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true}}, b.savedHousing)
}
func (b *backend) savedHousing(ctx context.Context, _ *mcp.CallToolRequest, q housingQuery) (*mcp.CallToolResult, housingOutput, error) {
	o := housingOutput{Source: "companion_save", Types: []housingType{}, Individuals: []housingPerson{}, Limit: 25, State: "missing", Note: "Latest ingested save only; ingestion and retrieval are not game capture time. Capacity uses parser-recognized player housing and template beds; deficit is not proven homeless people. Individual housed observations remain separate. Missing NPC ownership is excluded and disclosed. Null capacity or population means unknown. Lists bounded independently; no live fallback."}
	if q.Limit != nil {
		o.Limit = *q.Limit
	}
	if o.Limit < 1 || o.Limit > 100 {
		return nil, o, fmt.Errorf("limit must be 1 through 100")
	}
	_, raw, e := b.read(ctx, "/api/world", "companion_save", "")
	if e != nil {
		return nil, o, fmt.Errorf("saved world unavailable")
	}
	var w struct {
		SnapshotID *int64  `json:"snapshot_id"`
		IngestedAt *string `json:"ingested_at"`
		Housing    *struct {
			Quarters *int           `json:"quarters"`
			Houses   *int           `json:"houses"`
			Types    *[]housingType `json:"byType"`
		} `json:"housing"`
		NPCs *[]struct {
			Player *bool   `json:"is_player_npc"`
			GUID   *string `json:"guid"`
			Housed *bool   `json:"housed"`
		} `json:"npcs"`
	}
	if decodeEvidence(raw.Data, &w) != nil {
		return nil, o, fmt.Errorf("invalid housing evidence")
	}
	o.SnapshotID = w.SnapshotID
	o.IngestedAt = w.IngestedAt
	o.RetrievedAt = raw.RetrievedAt
	if w.Housing != nil {
		if w.Housing.Quarters == nil || w.Housing.Houses == nil {
			o.State = "unknown"
		} else {
			if *w.Housing.Quarters < 0 || *w.Housing.Houses < 0 {
				return nil, o, fmt.Errorf("negative housing capacity")
			}
			o.State = "available"
			o.Quarters = w.Housing.Quarters
			o.Houses = w.Housing.Houses
		}
		if w.Housing.Types != nil {
			sumBeds, sumHouses := 0, 0
			o.TypeTotal = len(*w.Housing.Types)
			for _, v := range *w.Housing.Types {
				if v.Class == "" || v.Capacity == nil || v.Count == nil || v.Beds == nil || *v.Capacity < 0 || *v.Count < 0 || *v.Beds < 0 {
					return nil, o, fmt.Errorf("invalid housing type evidence")
				}
				if *v.Count != 0 && *v.Capacity > int(^uint(0)>>1) / *v.Count {
					return nil, o, fmt.Errorf("housing capacity overflow")
				}
				if *v.Capacity**v.Count != *v.Beds {
					return nil, o, fmt.Errorf("housing type capacity mismatch")
				}
				if sumBeds > int(^uint(0)>>1)-*v.Beds || sumHouses > int(^uint(0)>>1)-*v.Count {
					return nil, o, fmt.Errorf("housing total overflow")
				}
				sumBeds += *v.Beds
				sumHouses += *v.Count
				if len(o.Types) < o.Limit {
					o.Types = append(o.Types, v)
				}
			}
			if o.Quarters != nil && (sumBeds != *o.Quarters || sumHouses != *o.Houses) {
				return nil, o, fmt.Errorf("housing totals mismatch")
			}
		}
	}
	if w.NPCs != nil {
		count := 0
		for _, n := range *w.NPCs {
			if n.Player == nil {
				o.UnknownOwnership++
				continue
			}
			if !*n.Player {
				continue
			}
			count++
			if n.Housed == nil {
				o.UnknownHousing++
			} else if *n.Housed {
				o.ObservedHoused++
			} else {
				o.ObservedNotHoused++
			}
			if len(o.Individuals) < o.Limit {
				o.Individuals = append(o.Individuals, housingPerson{GUID: n.GUID, Housed: n.Housed})
			}
		}
		o.Recruited = &count
		o.IndividualTotal = count
		if o.Quarters != nil {
			deficit, spare := 0, 0
			if count > *o.Quarters {
				deficit = count - *o.Quarters
			} else {
				spare = *o.Quarters - count
			}
			o.CapacityDeficit = &deficit
			o.SpareCapacity = &spare
		}
	}
	o.TypeReturned = len(o.Types)
	o.TypeTruncated = o.TypeTotal - o.TypeReturned
	o.IndividualReturned = len(o.Individuals)
	o.IndividualTruncated = o.IndividualTotal - o.IndividualReturned
	return nil, o, nil
}
