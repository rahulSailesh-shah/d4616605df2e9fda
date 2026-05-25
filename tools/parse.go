package tools

import (
	"encoding/json"
	"sort"
)

func Parse(data []byte) (typ, msg string, err error) {
	var in struct {
		Type    string          `json:"type"`
		Message json.RawMessage `json:"message"`
	}
	if err = json.Unmarshal(data, &in); err != nil {
		return
	}
	typ = in.Type

	var parts []struct {
		Word      string `json:"word"`
		Timestamp int    `json:"timestamp"`
	}
	if json.Unmarshal(in.Message, &parts) == nil && len(parts) > 0 {
		sort.Slice(parts, func(i, j int) bool { return parts[i].Timestamp < parts[j].Timestamp })
		for i, p := range parts {
			if i > 0 {
				msg += " "
			}
			msg += p.Word
		}
		return
	}
	json.Unmarshal(in.Message, &msg)
	return
}
