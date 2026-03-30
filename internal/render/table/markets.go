package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderMarkets(w io.Writer, res model.Response[[]model.CoinMarket]) error {
	printHeader(w, res)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "#\tID\tSYMBOL\tNAME\tPRICE\tCHANGE 24H\tMKT CAP\tVOLUME 24H")
	for _, m := range res.Data {
		rank := "-"
		if m.MarketCapRank != nil {
			rank = fmt.Sprintf("%d", *m.MarketCapRank)
		}
		chg := fmtOptF(m.PriceChangePct24h, "%.2f%%")
		cap := fmtOptF(m.MarketCap, "%.0f")
		vol := fmtOptF(m.Volume24h, "%.0f")
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%.6f\t%s\t%s\t%s\n",
			rank, m.ID, m.Symbol, m.Name, m.Price, chg, cap, vol)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}
