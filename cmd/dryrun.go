package cmd

import (
	"github.com/coingecko/coingecko-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().Bool("dry-run", false, "Preview API request without executing (JSON output)")
}

func isDryRun(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("dry-run")
	return v
}

type dryRunOutput struct {
	Method         string            `json:"method"`
	URL            string            `json:"url"`
	Params         map[string]string `json:"params"`
	Headers        map[string]string `json:"headers"`
	OASOperationID string            `json:"oas_operation_id,omitempty"`
	OASSpec        string            `json:"oas_spec,omitempty"`
	Pagination     *paginationInfo   `json:"pagination"`
}

type paginationInfo struct {
	TotalRequested int `json:"total_requested"`
	PerPage        int `json:"per_page"`
	Pages          int `json:"pages"`
}

func printDryRun(cfg *config.Config, cmdName, endpoint string, params map[string]string, pagination *paginationInfo) error {
	return printDryRunWithOp(cfg, cmdName, "", endpoint, params, pagination)
}

func printDryRunWithOp(cfg *config.Config, cmdName, opKey, endpoint string, params map[string]string, pagination *paginationInfo) error {
	headerKey, _ := cfg.AuthHeader()
	masked := cfg.MaskedKey()

	headers := map[string]string{}
	if cfg.APIKey != "" {
		headers[headerKey] = masked
	}
	headers["Accept"] = "application/json"

	out := dryRunOutput{
		Method:     "GET",
		URL:        cfg.BaseURL() + endpoint,
		Params:     params,
		Headers:    headers,
		Pagination: pagination,
	}

	if meta, ok := commandMeta[cmdName]; ok {
		out.OASSpec = meta.OASSpec
		if opKey != "" && meta.OASOperationIDs != nil {
			out.OASOperationID = meta.OASOperationIDs[opKey]
		} else {
			out.OASOperationID = meta.OASOperationID
		}
	}

	return printJSONRaw(out)
}
