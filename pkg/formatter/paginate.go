package formatter

import "fmt"

// FormatListWithContinue formats a list of items and optionally includes a continue token
// for pagination. For JSON, returns {"items": [...], "continue": "token"} when continue is set;
// otherwise just the items. For table format, returns the table and appends a pagination hint when continue is set.
func FormatListWithContinue(f Formatter, items interface{}, continueToken, format string) (string, error) {
	out, err := f.Format(items, format)
	if err != nil {
		return "", err
	}
	if continueToken == "" {
		return out, nil
	}
	switch format {
	case FormatJSON:
		// Re-format as wrapper object; the formatter already formatted items as JSON
		// We need the raw items for the wrapper. The Format() returns a string.
		// Simpler: build wrapper manually for JSON. We have items as interface{}.
		wrapper := map[string]interface{}{
			"items":    items,
			"continue": continueToken,
		}
		return f.Format(wrapper, format)
	case FormatTable:
		return out + fmt.Sprintf("\n\n(More available. Pass continue=%q for next page.)", continueToken), nil
	default:
		return out, nil
	}
}
