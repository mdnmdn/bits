package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
)

func RenderPrices(w io.Writer, res model.Response[[]model.CoinPrice]) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSYMBOL\tPRICE\tCURRENCY\tCHANGE 24H")
	for _, p := range res.Data {
		change := "-"
		if p.Change24h != nil {
			change = fmt.Sprintf("%.2f%%", *p.Change24h)
		}
		fmt.Fprintf(tw, "%s\t%s\t%.6f\t%s\t%s\n", p.ID, p.Symbol, p.Price, p.Currency, change)
	}
	tw.Flush()
	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintf(w, "\n%s\n", note)
	}
	return nil
}
