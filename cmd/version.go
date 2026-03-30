package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var version = "dev"
var commit = "none"
var date = "unknown"

var noCheckVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show bits version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("bits version %s\n", version)
		if commit != "none" {
			fmt.Printf("commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("built: %s\n", date)
		}

		noCheck := noCheckVersion || os.Getenv("BITS_NO_CHECK_VERSION") != "" || !checkVersionEnabled()
		if !noCheck {
			latest, _ := checkLatestVersion(context.Background())
			if latest != "" && latest != version {
				fmt.Printf("\nA new version is available: %s\n", latest)
			}
		}
	},
}

func checkVersionEnabled() bool {
	cfg, err := loadConfig()
	if err != nil {
		return true // default to checking
	}
	if cfg.CheckNewVersion == nil {
		return true // default to checking
	}
	return *cfg.CheckNewVersion
}

func init() {
	versionCmd.Flags().BoolVar(&noCheckVersion, "no-check", false, "Skip version check")
	RootCmd.AddCommand(versionCmd)
}

func checkLatestVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/mdnmdn/bits/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyStr := string(body)
	tagStart := strings.Index(bodyStr, `"tag_name":"`)
	if tagStart == -1 {
		return "", nil
	}
	tagStart += len(`"tag_name":"`)
	tagEnd := strings.Index(bodyStr[tagStart:], `"`)
	if tagEnd == -1 {
		return "", nil
	}

	return strings.TrimPrefix(bodyStr[tagStart:tagStart+tagEnd], "v"), nil
}
