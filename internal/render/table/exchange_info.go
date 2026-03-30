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
	_, _ = fmt.Fprintf(w, "Exchange : %s\n", info.ExchangeID)
	_, _ = fmt.Fprintf(w, "Market   : %s\n", info.Market)
	if info.ServerTime != nil {
		_, _ = fmt.Fprintf(w, "Time     : %s\n", info.ServerTime.Format("2006-01-02 15:04:05 UTC"))
	}
	_, _ = fmt.Fprintf(w, "Symbols  : %d\n\n", len(info.Symbols))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "SYMBOL\tBASE\tQUOTE\tSTATUS")
	for _, s := range info.Symbols {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", s.Symbol, s.BaseAsset, s.QuoteAsset, s.Status)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}
