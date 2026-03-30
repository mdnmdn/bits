package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderPrices(w io.Writer, res model.Response[[]model.CoinPrice]) error {
	printHeader(w, res)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "ID\tSYMBOL\tPRICE\tCURRENCY\tCHANGE 24H")
	for _, p := range res.Data {
		change := "-"
		if p.Change24h != nil {
			change = fmt.Sprintf("%.2f%%", *p.Change24h)
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%.6f\t%s\t%s\n", p.ID, NormalizeSymbol(p.Symbol), p.Price, p.Currency, change)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}
