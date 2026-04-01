package table

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/model"
)

func RenderTickers(w io.Writer, res model.Response[[]model.Ticker24h]) error {
	printHeader(w, res)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "SYMBOL\tMARKET\tLAST\tCHANGE%\tHIGH\tLOW\tVOLUME")
	for _, t := range res.Data {
		chg := fmtOptF(t.PriceChangePercent, "%.2f%%")
		hi := fmtOptF(t.HighPrice, "%.2f")
		lo := fmtOptF(t.LowPrice, "%.2f")
		vol := fmtOptF(t.Volume, "%.2f")
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%.2f\t%s\t%s\t%s\t%s\n",
			NormalizeSymbol(t.Symbol), t.Market, t.LastPrice, chg, hi, lo, vol)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}

func fmtOptF(v *float64, format string) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf(format, *v)
}
