package formatter

import "encoding/json"

// JSONFormatter outputs data as JSON.
type JSONFormatter struct{}

func (JSONFormatter) Format(data interface{}, _ string) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
