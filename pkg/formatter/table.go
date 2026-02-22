package formatter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// TableFormatter outputs slice of maps as an aligned table.
// For a single object, prints key-value pairs.
type TableFormatter struct{}

func (TableFormatter) Format(data interface{}, _ string) (string, error) {
	if data == nil {
		return "", nil
	}
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice {
		return formatSliceTable(v)
	}
	return formatObjectTable(data)
}

func formatSliceTable(v reflect.Value) (string, error) {
	if v.Len() == 0 {
		return "(no items)", nil
	}
	// First row defines columns
	first := v.Index(0).Interface()
	keys := mapKeys(first)
	if len(keys) == 0 {
		return "(no columns)", nil
	}
	var b strings.Builder
	// Header
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\t")
		}
		b.WriteString(k)
	}
	b.WriteString("\n")
	for i := 0; i < v.Len(); i++ {
		row := v.Index(i).Interface()
		m, ok := toMap(row)
		if !ok {
			b.WriteString(fmt.Sprint(row))
			b.WriteString("\n")
			continue
		}
		for j, k := range keys {
			if j > 0 {
				b.WriteString("\t")
			}
			b.WriteString(fmt.Sprint(m[k]))
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func formatObjectTable(data interface{}) (string, error) {
	m, ok := toMap(data)
	if !ok {
		return fmt.Sprint(data), nil
	}
	keys := mapKeys(data)
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("\t")
		b.WriteString(fmt.Sprint(m[k]))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func toMap(v interface{}) (map[string]interface{}, bool) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, false
	}
	return m, true
}

func mapKeys(v interface{}) []string {
	m, ok := toMap(v)
	if !ok {
		return nil
	}
	prefer := []string{"name", "namespace", "status", "phase", "node", "ip", "age"}
	seen := make(map[string]bool)
	var keys []string
	for _, k := range prefer {
		if _, ok := m[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	for k := range m {
		if !seen[k] {
			keys = append(keys, k)
		}
	}
	return keys
}
