package clioutput

import (
	"bytes"
	"encoding/json"
	"strconv"
	"text/tabwriter"
)

// renderTable uses the same projected items as JSON. Each pointer is a single
// column, even when its value is an object or array. JSON quoting prevents
// arbitrary labels or values from injecting terminal control characters.
func (r *CollectionRenderer) renderTable(items []json.RawMessage) ([]byte, error) {
	columns := r.columns
	var out bytes.Buffer
	table := tabwriter.NewWriter(&out, 0, 4, 2, ' ', 0)
	for i, column := range columns {
		if i != 0 {
			_, _ = table.Write([]byte("\t"))
		}
		label, err := json.Marshal(column)
		if err != nil {
			return nil, err
		}
		_, _ = table.Write(label)
	}
	_, _ = table.Write([]byte("\n"))
	for _, item := range items {
		for i, column := range columns {
			if i != 0 {
				_, _ = table.Write([]byte("\t"))
			}
			parts, err := pointerParts(column)
			if err != nil {
				return nil, err
			}
			value, err := tableCell(item, parts, r.item.protection, r.item.include, r.item.omit, r.item.projection == nil || len(r.item.projection.Include) == 0)
			if err != nil {
				return nil, err
			}
			if value != nil {
				var compact bytes.Buffer
				if err := json.Compact(&compact, value); err != nil {
					return nil, err
				}
				_, _ = table.Write(compact.Bytes())
			}
		}
		_, _ = table.Write([]byte("\n"))
	}
	if err := table.Flush(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// tableCell resolves the original pointer before applying omissions. Array
// compaction inside another selected column cannot move this cell's address.
func tableCell(raw json.RawMessage, parts []string, context *Protection, include, omit *mask, full bool) (json.RawMessage, error) {
	full = full || include.selectsAll()
	if omit.selectsAll() {
		return nil, nil
	}
	value := bytes.TrimSpace(raw)
	if len(value) == 0 {
		return nil, nil
	}
	if len(parts) == 0 {
		result, keep, err := project(value, context, include, omit, full, false, "")
		if !keep {
			return nil, err
		}
		return result, err
	}
	switch value[0] {
	case '{':
		var object map[string]json.RawMessage
		if err := json.Unmarshal(value, &object); err != nil {
			return nil, err
		}
		var child *Protection
		if context != nil {
			child = context.Members[parts[0]]
			for _, required := range context.Required {
				if required == parts[0] {
					full = true
				}
			}
		}
		return tableCell(object[parts[0]], parts[1:], child, include.child(parts[0]), omit.child(parts[0]), full)
	case '[':
		var array []json.RawMessage
		if err := json.Unmarshal(value, &array); err != nil {
			return nil, err
		}
		index, err := strconv.Atoi(parts[0])
		if err != nil || index < 0 || strconv.Itoa(index) != parts[0] || index >= len(array) {
			return nil, nil
		}
		var child *Protection
		if context != nil {
			child = context.Items
		}
		return tableCell(array[index], parts[1:], child, include.child(parts[0]), omit.child(parts[0]), full)
	default:
		return nil, nil
	}
}

// protectedColumn keeps embedded identity context visible in a single object
// cell. It promotes a selected descendant to its nearest protected container,
// while ordinary property objects retain the caller's exact pointer column.
func protectedColumn(path string, context *Protection) string {
	parts, _ := pointerParts(path)
	prefix := ""
	for i, part := range parts {
		if context == nil {
			return path
		}
		if context.Items != nil {
			context = context.Items
		} else {
			context = context.Members[part]
		}
		prefix = pointerChild(prefix, part)
		if context != nil && len(context.Required) != 0 && i < len(parts)-1 {
			return prefix
		}
	}
	return path
}
