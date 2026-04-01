package render

// Format is the output format identifier.
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
	FormatYAML     Format = "yaml"
	FormatToon     Format = "toon"
)

// ParseFormat converts a string to a Format, defaulting to FormatTable.
func ParseFormat(s string) Format {
	switch Format(s) {
	case FormatTable, FormatJSON, FormatMarkdown, FormatYAML, FormatToon:
		return Format(s)
	default:
		return FormatTable
	}
}
