package command

import (
	"fmt"
	"os"
	"time"

	"github.com/mdnmdn/bits"
	"github.com/mdnmdn/bits/model"
	rendertable "github.com/mdnmdn/bits/render/table"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Exchange symbol catalogue",
	RunE:  runInfo,
}

func init() {
	InfoCmd.Flags().String("symbol", "", "Filter by symbol")
	Root.AddCommand(InfoCmd)
}

func runInfo(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	symbolFilter, _ := cmd.Flags().GetString("symbol")

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	res, err := client.ExchangeInfo(cmd.Context(), market)
	if err != nil {
		return err
	}
	res.Market = market

	if symbolFilter != "" {
		filtered := res.Data.Symbols[:0]
		for _, s := range res.Data.Symbols {
			if s.Symbol == symbolFilter {
				filtered = append(filtered, s)
			}
		}
		res.Data.Symbols = filtered
	}

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderExchangeInfo(os.Stdout, res)
}

var TickerCmd = &cobra.Command{
	Use:   "ticker <symbol>...",
	Short: "24h rolling statistics",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runTicker,
}

func init() {
	Root.AddCommand(TickerCmd)
}

func runTicker(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	var results []model.Ticker24h
	for _, symbol := range args {
		res, err := client.Ticker24h(cmd.Context(), symbol, market)
		if err != nil {
			return err
		}
		results = append(results, res.Data)
	}

	res := model.Response[[]model.Ticker24h]{
		Kind:     model.KindTicker,
		Provider: client.ID(),
		Market:   market,
		Data:     results,
	}

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderTickers(os.Stdout, res)
}

var BookCmd = &cobra.Command{
	Use:   "book <symbol>",
	Short: "Order book snapshot",
	Args:  cobra.ExactArgs(1),
	RunE:  runBook,
}

func init() {
	BookCmd.Flags().Int("depth", 20, "Order book depth")
	Root.AddCommand(BookCmd)
}

func runBook(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	depth, _ := cmd.Flags().GetInt("depth")

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	res, err := client.OrderBook(cmd.Context(), args[0], market, depth)
	if err != nil {
		return err
	}
	res.Market = market

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderOrderBook(os.Stdout, res)
}

var CandlesCmd = &cobra.Command{
	Use:   "candles <symbol>",
	Short: "OHLCV history",
	Args:  cobra.ExactArgs(1),
	RunE:  runCandles,
}

func init() {
	CandlesCmd.Flags().String("interval", "1h", "Candle interval (e.g. 1m, 5m, 1h, 1d)")
	CandlesCmd.Flags().String("from", "", "Start time (RFC3339 or YYYY-MM-DD)")
	CandlesCmd.Flags().String("to", "", "End time (RFC3339 or YYYY-MM-DD)")
	CandlesCmd.Flags().Int("limit", 0, "Maximum number of candles (0 = provider default)")
	Root.AddCommand(CandlesCmd)
}

func runCandles(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	providerID, market, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	interval, _ := cmd.Flags().GetString("interval")
	fromStr, _ := cmd.Flags().GetString("from")
	toStr, _ := cmd.Flags().GetString("to")
	limit, _ := cmd.Flags().GetInt("limit")

	candleOpts := model.CandleOpts{}
	if fromStr != "" {
		t, terr := parseTime(fromStr)
		if terr != nil {
			return terr
		}
		candleOpts.From = &t
	}
	if toStr != "" {
		t, terr := parseTime(toStr)
		if terr != nil {
			return terr
		}
		candleOpts.To = &t
	}
	if limit > 0 {
		candleOpts.Limit = &limit
	}

	client := bits.NewProvider(cfg, providerID, bits.WithSymbolEngine())

	res, err := client.Candles(cmd.Context(), args[0], market, interval, candleOpts)
	if err != nil {
		return err
	}
	res.Market = market

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderCandles(os.Stdout, res)
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised time format %q (use RFC3339 or YYYY-MM-DD)", s)
}

var MarketsCmd = &cobra.Command{
	Use:   "markets",
	Short: "Ranked coin listing (aggregators only)",
	RunE:  runMarkets,
}

func init() {
	MarketsCmd.Flags().String("currency", "usd", "Quote currency")
	MarketsCmd.Flags().Int("page", 1, "Page number")
	MarketsCmd.Flags().Int("per-page", 100, "Results per page")
	MarketsCmd.Flags().String("order", "market_cap_desc", "Sort order")
	MarketsCmd.Flags().String("category", "", "Coin category filter")
	Root.AddCommand(MarketsCmd)
}

func runMarkets(cmd *cobra.Command, args []string) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	_, _, format, err := ResolveOptions(cmd)
	if err != nil {
		return err
	}

	currency, _ := cmd.Flags().GetString("currency")
	page, _ := cmd.Flags().GetInt("page")
	perPage, _ := cmd.Flags().GetInt("per-page")
	order, _ := cmd.Flags().GetString("order")
	category, _ := cmd.Flags().GetString("category")

	client := bits.NewProvider(cfg, "coingecko", bits.WithSymbolEngine())

	res, err := client.CoinMarkets(cmd.Context(), model.MarketOpts{
		Currency: currency,
		PerPage:  perPage,
		Page:     page,
		Order:    order,
		Category: category,
	})
	if err != nil {
		return err
	}

	if ok, err := RenderGeneric(os.Stdout, format, res); ok || err != nil {
		return err
	}
	return rendertable.RenderMarkets(os.Stdout, res)
}
