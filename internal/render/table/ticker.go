package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
)

func RenderTickers(w io.Writer, res model.Response[[]model.Ticker24h]) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SYMBOL\tMARKET\tLAST\tCHANGE%\tHIGH\tLOW\tVOLUME")
	for _, t := range res.Data {
		chg := fmtOptF(t.PriceChangePercent, "%.2f%%")
		hi := fmtOptF(t.HighPrice, "%.2f")
		lo := fmtOptF(t.LowPrice, "%.2f")
		vol := fmtOptF(t.Volume, "%.2f")
		fmt.Fprintf(tw, "%s\t%s\t%.2f\t%s\t%s\t%s\t%s\n",
			t.Symbol, t.Market, t.LastPrice, chg, hi, lo, vol)
	}
	tw.Flush()
	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintf(w, "\n%s\n", note)
	}
	return nil
}

func fmtOptF(v *float64, format string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf(format, *v)
}
