package formatter

// Formatter formats structured data for tool output.
type Formatter interface {
	Format(data interface{}, format string) (string, error)
}

// Supported formats.
const (
	FormatJSON  = "json"
	FormatTable = "table"
	FormatYAML  = "yaml"
)
