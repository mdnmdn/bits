// api-explorer - Interactive CLI tool for testing exchange provider APIs.
//
// Loads a TOML config file (_docs/providers/apis/api-examples.toml) that
// defines providers, endpoints, and named examples (parameter presets).
//
// ─── Modes ───────────────────────────────────────────────────────────────────
//
// 1. TUI (default, no args)
//    - Shows a flat, searchable list of all API calls across all providers.
//    - Type to filter: space-separated terms are AND-matched against provider
//      name, endpoint name, category, and example label.
//    - Navigate with ↑/↓ or j/k.
//    - Press Enter to run the selected API call, print the JSON response, and exit.
//    - Press q or Ctrl+C to quit.
//
// 2. List (first arg = "list")
//    go run api-explorer.go list
//    - Prints all entries grouped by provider to stdout.
//    - Each line shows: <key>  <description>
//    - The key format is: <provider>.<endpoint-slug>[.<example-slug>]
//    - Example output:
//        binance.price-ticker.btcusdt    Latest price for BTC/USDT
//        coingecko.simple-price.btc-eth  Current prices for BTC and ETH
//
// 3. Run by key (first arg != "list" and != "--help")
//    go run api-explorer.go <key>
//    - Looks up the entry by its dot-separated key.
//    - Executes the HTTP request with the example's parameters.
//    - Prints a header (name, URL, status, latency) followed by pretty-printed
//      JSON response to stdout, then exits.
//    - If the key is not found, prints an error to stderr and exits with code 1.
//
// ─── Key Format ──────────────────────────────────────────────────────────────
//
//    <provider-id>.<endpoint-name-slug>[.<example-label-slug>]
//
//    Slugs are lowercased, spaces become hyphens, slashes/parens are stripped.
//    Examples:
//      binance.server-time
//      binance.price-ticker.btcusdt
//      binance.klines.btc-1d-last-30
//      coingecko.simple-price.btc-eth-in-usd
//      mexc.futures-klines.btc-4h
//
// ─── Config Format (TOML) ────────────────────────────────────────────────────
//
//    [providers.<id>]
//    name       = "Display Name"
//    base_url   = "https://api.example.com"
//    api_key_env = "PROVIDER_API_KEY"   # optional, read from env at runtime
//
//      [[providers.<id>.endpoints]]
//      name        = "Endpoint Name"
//      category    = "market"           # market | general | futures
//      method      = "GET"
//      path        = "/api/v1/ticker"
//      description = "Short description"
//      params      = ["symbol"]         # query param names
//      path_params = ["id"]             # path param names (replaced in {id})
//      defaults    = { symbol = "BTC" } # default values for params
//
//        [[providers.<id>.endpoints.examples]]
//        label  = "BTC/USDT"
//        params = { symbol = "BTCUSDT" }
//
//    If an endpoint has no examples, a single entry is created using defaults.
//    If an endpoint has N examples, N entries are created (one per example).
//
// ─── HTTP Behavior ───────────────────────────────────────────────────────────
//
//    - Path params ({name}) are replaced in the URL path before query encoding.
//    - Query params are URL-encoded and appended.
//    - Empty-string params are omitted from the query.
//    - If api_key_env is set and the env var exists, two headers are added:
//        Authorization: Bearer <key>
//        X-API-Key: <key>
//    - Timeout: 30 seconds.
//    - Response body is pretty-printed JSON (falls back to raw if invalid).
//
// ─── Output ──────────────────────────────────────────────────────────────────
//
//    Success:
//      <Endpoint Name> — <Example Label>
//      <full URL>
//      200 OK  150ms
//
//      { ... pretty JSON ... }
//
//    Error (non-2xx):
//      <Endpoint Name> — <Example Label>
//      <full URL>
//      400  80ms
//
//      { ... error JSON from server ... }
//
//    Network/parse error:
//      Error: <message>   (to stderr, exit 1)

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Config Types ────────────────────────────────────────────────────────────

type Example struct {
	Label  string            `toml:"label"`
	Params map[string]string `toml:"params"`
}

type Endpoint struct {
	Name        string            `toml:"name"`
	Category    string            `toml:"category"`
	Method      string            `toml:"method"`
	Path        string            `toml:"path"`
	Description string            `toml:"description"`
	Params      []string          `toml:"params"`
	PathParams  []string          `toml:"path_params"`
	Defaults    map[string]string `toml:"defaults"`
	Examples    []Example         `toml:"examples"`
}

type Provider struct {
	Name      string     `toml:"name"`
	BaseURL   string     `toml:"base_url"`
	APIKeyEnv string     `toml:"api_key_env"`
	Endpoints []Endpoint `toml:"endpoints"`
}

type Config struct {
	Providers map[string]Provider `toml:"providers"`
}

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	urlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	searchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("236")).Padding(0, 1)
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	providerTag   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	catTag        = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Background(lipgloss.Color("57"))
)

// ─── Flat Entry ──────────────────────────────────────────────────────────────

type entry struct {
	providerID string
	provider   Provider
	endpoint   Endpoint
	example    Example
	params     map[string]string
}

func (e entry) searchText() string {
	return strings.ToLower(e.provider.Name + " " + e.providerID + " " + e.endpoint.Name + " " + e.endpoint.Description + " " + e.example.Label + " " + e.endpoint.Category)
}

func (e entry) displayLine(w int) string {
	tag := catTag.Render("[" + e.endpoint.Category + "]")
	prov := providerTag.Render(e.provider.Name)
	name := e.endpoint.Name
	label := e.example.Label
	params := buildParamDesc(e.example.Params, e.endpoint)

	line := fmt.Sprintf("  %s  %s  %s", tag, prov, name)
	if label != "" {
		line += fmt.Sprintf("  — %s", label)
	}
	if params != "" {
		line += "  " + dimStyle.Render(params)
	}

	if w > 2 && len(line) > w-2 {
		line = line[:w-2]
	}
	return line
}

// ─── Key / Slug ──────────────────────────────────────────────────────────────

func entryKey(e entry) string {
	parts := []string{e.providerID, slug(e.endpoint.Name)}
	if e.example.Label != "" {
		parts = append(parts, slug(e.example.Label))
	}
	return strings.Join(parts, ".")
}

func slug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	return s
}

// ─── Model ───────────────────────────────────────────────────────────────────

type model struct {
	config   Config
	entries  []entry
	filtered []int
	cursor   int
	search   string
	width    int
	height   int
	selected int
	done     bool
}

func initialModel(cfg Config) model {
	providerIDs := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		providerIDs = append(providerIDs, id)
	}
	sort.Strings(providerIDs)

	var entries []entry
	for _, pid := range providerIDs {
		p := cfg.Providers[pid]
		for _, ep := range p.Endpoints {
			if len(ep.Examples) == 0 {
				params := buildParamsFromDefaults(ep)
				entries = append(entries, entry{
					providerID: pid,
					provider:   p,
					endpoint:   ep,
					example:    Example{},
					params:     params,
				})
			} else {
				for _, ex := range ep.Examples {
					params := buildParamsFromExample(ep, ex)
					entries = append(entries, entry{
						providerID: pid,
						provider:   p,
						endpoint:   ep,
						example:    ex,
						params:     params,
					})
				}
			}
		}
	}

	m := model{
		config:   cfg,
		entries:  entries,
		filtered: make([]int, len(entries)),
		cursor:   0,
	}
	for i := range entries {
		m.filtered[i] = i
	}
	return m
}

// ─── Search ──────────────────────────────────────────────────────────────────

func (m *model) applyFilter() {
	m.filtered = m.filtered[:0]
	terms := strings.Fields(strings.ToLower(m.search))
	for i, e := range m.entries {
		if len(terms) == 0 {
			m.filtered = append(m.filtered, i)
			continue
		}
		text := e.searchText()
		match := true
		for _, t := range terms {
			if !strings.Contains(text, t) {
				match = false
				break
			}
		}
		if match {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		if len(m.filtered) > 0 {
			m.cursor = len(m.filtered) - 1
		} else {
			m.cursor = 0
		}
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func buildParamsFromDefaults(ep Endpoint) map[string]string {
	params := make(map[string]string)
	all := append([]string{}, ep.Params...)
	all = append(all, ep.PathParams...)
	for _, p := range all {
		if def, ok := ep.Defaults[p]; ok {
			params[p] = def
		} else {
			params[p] = ""
		}
	}
	return params
}

func buildParamsFromExample(ep Endpoint, ex Example) map[string]string {
	params := make(map[string]string)
	all := append([]string{}, ep.Params...)
	all = append(all, ep.PathParams...)
	for _, p := range all {
		if v, ok := ex.Params[p]; ok {
			params[p] = v
		} else if def, ok := ep.Defaults[p]; ok {
			params[p] = def
		} else {
			params[p] = ""
		}
	}
	return params
}

func buildParamDesc(params map[string]string, ep Endpoint) string {
	if len(params) == 0 {
		return ""
	}
	var parts []string
	for _, p := range ep.Params {
		if v, ok := params[p]; ok && v != "" {
			parts = append(parts, p+"="+v)
		}
	}
	for _, p := range ep.PathParams {
		if v, ok := params[p]; ok && v != "" {
			parts = append(parts, p+"="+v)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "  ")
}

// ─── Fetch & Print ───────────────────────────────────────────────────────────

func runEntry(e entry) {
	apiKey := ""
	if e.provider.APIKeyEnv != "" {
		apiKey = os.Getenv(e.provider.APIKeyEnv)
	}

	fullPath := e.endpoint.Path
	for _, pp := range e.endpoint.PathParams {
		if val, ok := e.params[pp]; ok {
			fullPath = strings.Replace(fullPath, "{"+pp+"}", val, 1)
		}
	}
	query := url.Values{}
	for _, p := range e.endpoint.Params {
		if val, ok := e.params[p]; ok && val != "" {
			query.Set(p, val)
		}
	}
	if len(query) > 0 {
		fullPath += "?" + query.Encode()
	}
	fullURL := e.provider.BaseURL + fullPath

	start := time.Now()
	req, err := http.NewRequest(e.endpoint.Method, fullURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Accept", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	label := e.endpoint.Name
	if e.example.Label != "" {
		label += " — " + e.example.Label
	}

	status := ""
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = successStyle.Render(fmt.Sprintf("%d OK", resp.StatusCode))
	} else {
		status = errorStyle.Render(fmt.Sprintf("%d", resp.StatusCode))
	}

	fmt.Println(titleStyle.Render(label))
	fmt.Println(urlStyle.Render(fullURL))
	fmt.Println(statusStyle.Render(fmt.Sprintf("%s  %s", status, time.Since(start).Round(time.Millisecond))))
	fmt.Println()

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err == nil {
		fmt.Println(pretty.String())
	} else {
		fmt.Println(string(body))
	}
}

// ─── CLI: list ───────────────────────────────────────────────────────────────

func printList(cfg Config) {
	providerIDs := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		providerIDs = append(providerIDs, id)
	}
	sort.Strings(providerIDs)

	for _, pid := range providerIDs {
		p := cfg.Providers[pid]
		fmt.Printf("\n%s (%s)\n", p.Name, pid)
		fmt.Println(strings.Repeat("─", 60))
		for _, ep := range p.Endpoints {
			if len(ep.Examples) == 0 {
				params := buildParamsFromDefaults(ep)
				e := entry{providerID: pid, provider: p, endpoint: ep, params: params}
				fmt.Printf("  %-45s %s\n", entryKey(e), ep.Description)
			} else {
				for _, ex := range ep.Examples {
					params := buildParamsFromExample(ep, ex)
					e := entry{providerID: pid, provider: p, endpoint: ep, example: ex, params: params}
					label := ex.Label
					if label == "" {
						label = ep.Description
					}
					fmt.Printf("  %-45s %s\n", entryKey(e), label)
				}
			}
		}
	}
}

// ─── CLI: run by key ─────────────────────────────────────────────────────────

func runByKey(cfg Config, key string) {
	providerIDs := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		providerIDs = append(providerIDs, id)
	}
	sort.Strings(providerIDs)

	for _, pid := range providerIDs {
		p := cfg.Providers[pid]
		for _, ep := range p.Endpoints {
			if len(ep.Examples) == 0 {
				params := buildParamsFromDefaults(ep)
				e := entry{providerID: pid, provider: p, endpoint: ep, params: params}
				if entryKey(e) == key {
					runEntry(e)
					return
				}
			} else {
				for _, ex := range ep.Examples {
					params := buildParamsFromExample(ep, ex)
					e := entry{providerID: pid, provider: p, endpoint: ep, example: ex, params: params}
					if entryKey(e) == key {
						runEntry(e)
						return
					}
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "Error: no entry found for key %q\n", key)
	fmt.Fprintln(os.Stderr, "Run with 'list' to see all available entries.")
	os.Exit(1)
}

// ─── Bubbletea: Init ─────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return tea.WindowSize()
}

// ─── Bubbletea: Update ───────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if len(m.filtered) > 0 {
				m.selected = m.filtered[m.cursor]
				m.done = true
				return m, tea.Quit
			}
		case "backspace":
			if len(m.search) > 0 {
				m.search = m.search[:len(m.search)-1]
				m.applyFilter()
			}
		default:
			if len(msg.String()) == 1 {
				m.search += msg.String()
				m.applyFilter()
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

// ─── Bubbletea: View ─────────────────────────────────────────────────────────

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" API Explorer ") + "  " + subtitleStyle.Render(fmt.Sprintf("%d/%d", len(m.filtered), len(m.entries))))
	b.WriteString("\n")

	searchDisplay := m.search
	if m.search == "" {
		searchDisplay = "type to search..."
		b.WriteString(searchStyle.Render(" / " + dimStyle.Render(searchDisplay)))
	} else {
		b.WriteString(searchStyle.Render(" / " + cursorStyle.Render(searchDisplay) + "█"))
	}
	b.WriteString("\n\n")

	maxItems := m.height - 6
	if maxItems < 1 {
		maxItems = 1
	}

	start := 0
	if m.cursor >= maxItems {
		start = m.cursor - maxItems + 1
	}
	end := start + maxItems
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	if len(m.filtered) == 0 {
		b.WriteString(dimStyle.Render("  no matches"))
	} else {
		for i := start; i < end; i++ {
			idx := m.filtered[i]
			e := m.entries[idx]
			line := e.displayLine(m.width)
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("▸" + line[1:]))
			} else {
				b.WriteString(line)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("↑↓/jk: navigate  enter: run & exit  q: quit"))
	return b.String()
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		printHelp()
		return
	}

	cfg, err := loadConfig("_docs/providers/apis/api-examples.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			printList(cfg)
			return
		default:
			runByKey(cfg, os.Args[1])
			return
		}
	}

	m := initialModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(model); ok && fm.done {
		runEntry(fm.entries[fm.selected])
	}
}

func printHelp() {
	fmt.Println(`API Explorer - Interactive API testing tool

Usage:
  go run api-explorer.go              # Interactive TUI mode
  go run api-explorer.go list         # List all available API calls
  go run api-explorer.go <key>        # Run a specific API call

Examples:
  go run api-explorer.go list
  go run api-explorer.go binance.price-ticker.btcusdt
  go run api-explorer.go coingecko.simple-price.btc-eth-in-usd
  go run api-explorer.go mexc.klines.btc-1h

Environment Variables:
  BINANCE_API_KEY       Binance API key
  BITGET_API_KEY        Bitget API key
  COINGECKO_API_KEY     CoinGecko API key
  WHITEBIT_API_KEY      WhiteBit API key
  CRYPTOCOM_API_KEY     Crypto.com API key
  MEXC_API_KEY          MEXC API key`)
}

func loadConfig(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
