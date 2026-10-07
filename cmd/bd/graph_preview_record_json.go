package main

import (
	"bytes"
	"encoding/json"
	"strings"
)

// graphProjectCompleteRecords changes only the CLI representation. The store
// still reads and writes its original snapshots, including their version field.
func graphProjectCompleteRecords(result any) (json.RawMessage, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return graphProjectRecordJSON(raw)
}

func graphProjectRecordJSON(raw json.RawMessage) (json.RawMessage, error) {
	return graphProjectRecordJSONIn(raw, false)
}

func graphProjectRecordJSONIn(raw json.RawMessage, ownedComparisonValue bool) (json.RawMessage, error) {
	value := bytes.TrimSpace(raw)
	if len(value) == 0 {
		return raw, nil
	}
	switch value[0] {
	case '{':
		var members map[string]json.RawMessage
		if err := json.Unmarshal(value, &members); err != nil {
			return nil, err
		}
		if graphIsCompleteRecord(members) {
			delete(members, "version")
		}
		ownedChange := string(members["area"]) == `"owned"`
		for key, member := range members {
			// These are caller-authored values, not graph record wrappers. An
			// object inside them may legitimately have a member named version.
			switch key {
			case "properties", "metadata", "attribution":
				continue
			case "value":
				if !ownedComparisonValue {
					continue
				}
			}
			projected, err := graphProjectRecordJSONIn(member, ownedComparisonValue || (ownedChange && (key == "from" || key == "to")))
			if err != nil {
				return nil, err
			}
			members[key] = projected
		}
		return json.Marshal(members)
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(value, &items); err != nil {
			return nil, err
		}
		for i, item := range items {
			projected, err := graphProjectRecordJSONIn(item, ownedComparisonValue)
			if err != nil {
				return nil, err
			}
			items[i] = projected
		}
		return json.Marshal(items)
	default:
		return raw, nil
	}
}

func graphIsCompleteRecord(members map[string]json.RawMessage) bool {
	for _, key := range []string{"id", "type", "revision", "version", "properties"} {
		if _, present := members[key]; !present {
			return false
		}
	}
	var id string
	if err := json.Unmarshal(members["id"], &id); err != nil {
		return false
	}
	return (strings.HasPrefix(id, "https://") || strings.HasPrefix(id, "http://")) &&
		(strings.Contains(id, "/beads/") || strings.Contains(id, "/links/"))
}
