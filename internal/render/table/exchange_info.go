package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderExchangeInfo(w io.Writer, res model.Response[model.ExchangeInfo]) error {
	printHeader(w, res)
	info := res.Data
	fmt.Fprintf(w, "Exchange : %s\n", info.ExchangeID)
	fmt.Fprintf(w, "Market   : %s\n", info.Market)
	if info.ServerTime != nil {
		fmt.Fprintf(w, "Time     : %s\n", info.ServerTime.Format("2006-01-02 15:04:05 UTC"))
	}
	fmt.Fprintf(w, "Symbols  : %d\n\n", len(info.Symbols))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SYMBOL\tBASE\tQUOTE\tSTATUS")
	for _, s := range info.Symbols {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Symbol, s.BaseAsset, s.QuoteAsset, s.Status)
	}
	tw.Flush()
	printFooter(w, res)
	return nil
}
