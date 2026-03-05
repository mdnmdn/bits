package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const oasRepo = "https://github.com/coingecko/coingecko-api-oas"

type commandAnnotation struct {
	APIEndpoint    string            `json:"api_endpoint,omitempty"`
	APIEndpoints   map[string]string `json:"api_endpoints,omitempty"`
	OASOperationID string            `json:"oas_operation_id,omitempty"`
	OASOperationIDs map[string]string `json:"oas_operation_ids,omitempty"`
	OASSpec        string            `json:"oas_spec,omitempty"`
	PaidOnly       bool              `json:"paid_only"`
	RequiresAuth   bool              `json:"requires_auth"`
}

var commandMeta = map[string]commandAnnotation{
	"price": {
		APIEndpoint:    "/simple/price",
		OASOperationID: "simple-price",
		OASSpec:        "coingecko-demo.json",
		RequiresAuth:   true,
	},
	"markets": {
		APIEndpoint:    "/coins/markets",
		OASOperationID: "coins-markets",
		OASSpec:        "coingecko-demo.json",
		RequiresAuth:   true,
	},
	"search": {
		APIEndpoint:    "/search",
		OASOperationID: "search-data",
		OASSpec:        "coingecko-demo.json",
		RequiresAuth:   true,
	},
	"trending": {
		APIEndpoint:    "/search/trending",
		OASOperationID: "trending-search",
		OASSpec:        "coingecko-demo.json",
		RequiresAuth:   true,
	},
	"history": {
		APIEndpoints: map[string]string{
			"--date":      "/coins/{id}/history",
			"--days":      "/coins/{id}/ohlc",
			"--from/--to": "/coins/{id}/market_chart/range",
		},
		OASOperationIDs: map[string]string{
			"--date":      "coins-id-history",
			"--days":      "coins-id-ohlc",
			"--from/--to": "coins-id-market-chart-range",
		},
		OASSpec:      "coingecko-demo.json",
		RequiresAuth: true,
	},
	"top-gainers-losers": {
		APIEndpoint:    "/coins/top_gainers_losers",
		OASOperationID: "coins-top-gainers-losers",
		OASSpec:        "coingecko-pro.json",
		PaidOnly:       true,
		RequiresAuth:   true,
	},
}

type flagInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Default     string   `json:"default"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// Flag enums sourced from CoinGecko OAS spec.
var flagEnums = map[string]map[string][]string{
	"history": {
		"days":     {"1", "7", "14", "30", "90", "180", "365", "max"},
		"interval": {"5m", "hourly", "daily"},
	},
	"top-gainers-losers": {
		"duration":  {"1h", "24h", "7d", "14d", "30d", "60d", "1y"},
		"top-coins": {"300", "500", "1000", "all"},
	},
	"markets": {
		"order": {"market_cap_desc", "market_cap_asc", "volume_asc", "volume_desc", "id_asc", "id_desc"},
	},
}

type commandInfo struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Flags           []flagInfo        `json:"flags"`
	Examples        []string          `json:"examples,omitempty"`
	OutputFormats   []string          `json:"output_formats"`
	RequiresAuth    bool              `json:"requires_auth"`
	PaidOnly        bool              `json:"paid_only"`
	APIEndpoint     string            `json:"api_endpoint,omitempty"`
	APIEndpoints    map[string]string `json:"api_endpoints,omitempty"`
	OASOperationID  string            `json:"oas_operation_id,omitempty"`
	OASOperationIDs map[string]string `json:"oas_operation_ids,omitempty"`
}

type commandCatalog struct {
	Version  string        `json:"version"`
	OASRepo  string        `json:"oas_repo"`
	Commands []commandInfo `json:"commands"`
}

var commandsCmd = &cobra.Command{
	Use:    "commands",
	Short:  "Output machine-readable command catalog (JSON)",
	Hidden: true,
	RunE:   runCommands,
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}

func runCommands(cmd *cobra.Command, args []string) error {
	catalog := commandCatalog{
		Version: version,
		OASRepo: oasRepo,
	}

	for _, c := range rootCmd.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}

		// Skip non-data commands.
		if c.Name() == "auth" || c.Name() == "status" || c.Name() == "version" {
			info := commandInfo{
				Name:          c.Name(),
				Description:   c.Short,
				Flags:         extractFlags(c),
				Examples:      splitExamples(c.Example),
				OutputFormats: []string{"table"},
			}
			if c.Name() == "version" {
				info.OutputFormats = []string{"table", "json"}
			}
			catalog.Commands = append(catalog.Commands, info)
			continue
		}

		// Handle subcommands (tui markets, tui trending).
		if c.HasSubCommands() {
			for _, sub := range c.Commands() {
				if sub.Hidden {
					continue
				}
				info := commandInfo{
					Name:          c.Name() + " " + sub.Name(),
					Description:   sub.Short,
					Flags:         extractFlags(sub),
					Examples:      splitExamples(sub.Example),
					OutputFormats: []string{"tui"},
					RequiresAuth:  true,
				}
				catalog.Commands = append(catalog.Commands, info)
			}
			continue
		}

		info := commandInfo{
			Name:          c.Name(),
			Description:   c.Short,
			Flags:         extractFlags(c),
			Examples:      splitExamples(c.Example),
			OutputFormats: []string{"table", "json"},
			RequiresAuth:  true,
		}

		if meta, ok := commandMeta[c.Name()]; ok {
			info.PaidOnly = meta.PaidOnly
			info.RequiresAuth = meta.RequiresAuth
			info.APIEndpoint = meta.APIEndpoint
			info.APIEndpoints = meta.APIEndpoints
			info.OASOperationID = meta.OASOperationID
			info.OASOperationIDs = meta.OASOperationIDs
		}

		catalog.Commands = append(catalog.Commands, info)
	}

	return printJSONRaw(catalog)
}

func extractFlags(cmd *cobra.Command) []flagInfo {
	var flags []flagInfo
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		fi := flagInfo{
			Name:        f.Name,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		}
		if enums, ok := flagEnums[cmd.Name()]; ok {
			if e, ok := enums[f.Name]; ok {
				fi.Enum = e
			}
		}
		flags = append(flags, fi)
	})
	return flags
}

func splitExamples(s string) []string {
	if s == "" {
		return nil
	}
	var examples []string
	for _, line := range splitLines(s) {
		line = trimLeadingSpaces(line)
		if line != "" {
			examples = append(examples, line)
		}
	}
	return examples
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimLeadingSpaces(s string) string {
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	return s[i:]
}
