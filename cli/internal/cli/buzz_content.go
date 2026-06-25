// Shared helpers for the Agilix Buzz content-review commands (content tree,
// content diff). These read a course manifest via GetItemList and parse Buzz's
// {"$value": ...} property-bag item data.
//
// Hand-authored.
package cli

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"agilix-buzz-pp-cli/internal/client"
)

type contentItem struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	Parent    string  `json:"parent,omitempty"`
	Sequence  string  `json:"sequence,omitempty"`
	MaxPoints float64 `json:"max_points"`
	Gradable  bool    `json:"gradable"`
}

// dvalue unwraps Buzz's {"$value": X} property-bag wrapper.
func dvalue(raw json.RawMessage) json.RawMessage {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	if v, ok := m["$value"]; ok {
		return v
	}
	return nil
}

func dvStr(data map[string]json.RawMessage, key string) string {
	v := dvalue(data[key])
	if v == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return strings.Trim(string(v), `"`)
}

func dvNum(data map[string]json.RawMessage, key string) float64 {
	v := dvalue(data[key])
	if v == nil {
		return 0
	}
	m := map[string]json.RawMessage{"x": v}
	return rawNum(m, "x")
}

// fetchItems pulls a course's manifest items via GetItemList.
func fetchItems(ctx context.Context, c *client.Client, entityID string) ([]contentItem, error) {
	body, err := c.Get(ctx, "/cmd?cmd=getitemlist", map[string]string{"entityid": entityID})
	if err != nil {
		return nil, err
	}
	raws, err := client.DLAPList(body, "items", "item")
	if err != nil {
		return nil, err
	}
	out := make([]contentItem, 0, len(raws))
	for _, r := range raws {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(r, &m); err != nil {
			continue
		}
		ci := contentItem{ID: rawStr(m, "id")}
		if data := rawObj(m, "data"); data != nil {
			ci.Title = dvStr(data, "title")
			ci.Type = dvStr(data, "type")
			ci.Parent = dvStr(data, "parent")
			ci.Sequence = dvStr(data, "sequence")
			ci.MaxPoints = dvNum(data, "maxpoints")
			ci.Gradable = strings.EqualFold(dvStr(data, "gradable"), "true") || dvStr(data, "gradable") == "1"
		}
		if ci.Type == "" {
			ci.Type = "Folder"
		}
		out = append(out, ci)
	}
	return out, nil
}

// contentSummary rolls up counts and points across an item set.
type contentSummary struct {
	TotalItems  int            `json:"total_items"`
	Gradable    int            `json:"gradable_items"`
	TotalPoints float64        `json:"total_points"`
	ByType      map[string]int `json:"by_type"`
}

func summarize(items []contentItem) contentSummary {
	s := contentSummary{ByType: map[string]int{}}
	for _, it := range items {
		s.TotalItems++
		s.ByType[it.Type]++
		if it.Gradable {
			s.Gradable++
		}
		s.TotalPoints += it.MaxPoints
	}
	return s
}

// sortItemsBySeq sorts a slice of items by their sequence string (Buzz uses
// lexically-ordered sequence keys like "a", "b", "c").
func sortItemsBySeq(items []contentItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sequence != items[j].Sequence {
			return items[i].Sequence < items[j].Sequence
		}
		return items[i].Title < items[j].Title
	})
}
