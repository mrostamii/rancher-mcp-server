package formatter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormatter_Format(t *testing.T) {
	f := JSONFormatter{}

	t.Run("object", func(t *testing.T) {
		data := map[string]interface{}{"name": "foo", "count": 42}
		out, err := f.Format(data, FormatJSON)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if parsed["name"] != "foo" || parsed["count"] != float64(42) {
			t.Errorf("unexpected output: %s", out)
		}
		if !strings.Contains(out, "  ") {
			t.Error("expected indented JSON")
		}
	})

	t.Run("slice", func(t *testing.T) {
		data := []map[string]interface{}{{"a": 1}, {"b": 2}}
		out, err := f.Format(data, FormatJSON)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		var parsed []map[string]interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if len(parsed) != 2 {
			t.Errorf("expected 2 items, got %d", len(parsed))
		}
	})

	t.Run("nil", func(t *testing.T) {
		out, err := f.Format(nil, FormatJSON)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		if out != "null" {
			t.Errorf("expected null, got %q", out)
		}
	})
}

func TestTableFormatter_Format(t *testing.T) {
	f := TableFormatter{}

	t.Run("nil", func(t *testing.T) {
		out, err := f.Format(nil, FormatTable)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		if out != "" {
			t.Errorf("expected empty string, got %q", out)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		data := []map[string]interface{}{}
		out, err := f.Format(data, FormatTable)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		if out != "(no items)" {
			t.Errorf("expected (no items), got %q", out)
		}
	})

	t.Run("slice of maps", func(t *testing.T) {
		data := []map[string]interface{}{
			{"name": "a", "namespace": "ns1", "status": "Running"},
			{"name": "b", "namespace": "ns2", "status": "Pending"},
		}
		out, err := f.Format(data, FormatTable)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		lines := strings.Split(out, "\n")
		if len(lines) < 3 {
			t.Fatalf("expected header + 2 rows, got %d lines", len(lines))
		}
		if !strings.Contains(lines[0], "name") || !strings.Contains(lines[0], "namespace") {
			t.Errorf("header should contain name, namespace: %s", lines[0])
		}
		if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
			t.Errorf("output should contain row data: %s", out)
		}
		if !strings.Contains(out, "\t") {
			t.Error("expected tab-separated columns")
		}
	})

	t.Run("single map", func(t *testing.T) {
		data := map[string]interface{}{"name": "foo", "status": "ok"}
		out, err := f.Format(data, FormatTable)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		if !strings.Contains(out, "name") || !strings.Contains(out, "foo") {
			t.Errorf("expected key-value rows: %s", out)
		}
	})

	t.Run("preferred column order", func(t *testing.T) {
		data := []map[string]interface{}{
			{"z": 1, "name": "n", "namespace": "ns", "status": "s"},
		}
		out, err := f.Format(data, FormatTable)
		if err != nil {
			t.Fatalf("Format: %v", err)
		}
		// Preferred keys (name, namespace, status) should come before z
		nameIdx := strings.Index(out, "name")
		zIdx := strings.Index(out, "z")
		if nameIdx < 0 || zIdx < 0 {
			t.Fatalf("expected both name and z in output: %s", out)
		}
		if nameIdx > zIdx {
			t.Errorf("preferred keys should come first; name at %d, z at %d", nameIdx, zIdx)
		}
	})
}

func TestFormatListWithContinue(t *testing.T) {
	f := JSONFormatter{}
	items := []map[string]interface{}{{"name": "a"}, {"name": "b"}}

	t.Run("no continue token", func(t *testing.T) {
		out, err := FormatListWithContinue(f, items, "", FormatJSON)
		if err != nil {
			t.Fatalf("FormatListWithContinue: %v", err)
		}
		var parsed []map[string]interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if len(parsed) != 2 {
			t.Errorf("expected 2 items, got %d", len(parsed))
		}
	})

	t.Run("with continue token JSON", func(t *testing.T) {
		out, err := FormatListWithContinue(f, items, "next-page-token", FormatJSON)
		if err != nil {
			t.Fatalf("FormatListWithContinue: %v", err)
		}
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			t.Fatalf("output not valid JSON: %v", err)
		}
		if parsed["continue"] != "next-page-token" {
			t.Errorf("expected continue in output: %s", out)
		}
		it, ok := parsed["items"].([]interface{})
		if !ok || len(it) != 2 {
			t.Errorf("expected items array with 2 elements: %s", out)
		}
	})

	t.Run("with continue token table", func(t *testing.T) {
		tf := TableFormatter{}
		out, err := FormatListWithContinue(tf, items, "token123", FormatTable)
		if err != nil {
			t.Fatalf("FormatListWithContinue: %v", err)
		}
		if !strings.Contains(out, "token123") {
			t.Errorf("expected pagination hint with token: %s", out)
		}
		if !strings.Contains(out, "More available") {
			t.Errorf("expected pagination hint: %s", out)
		}
	})
}
