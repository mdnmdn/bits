package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderOrderBook(w io.Writer, res model.Response[model.OrderBook]) error {
	printHeader(w, res)
	ob := res.Data
	_, _ = fmt.Fprintf(w, "Order Book: %s (%s)\n\n", NormalizeSymbol(ob.Symbol), ob.Market)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "BIDS\t\tASKS")
	_, _ = fmt.Fprintln(tw, "PRICE\tQTY\tPRICE\tQTY")
	n := len(ob.Bids)
	if len(ob.Asks) > n {
		n = len(ob.Asks)
	}
	for i := 0; i < n; i++ {
		bidP, bidQ := "-", "-"
		askP, askQ := "-", "-"
		if i < len(ob.Bids) {
			bidP = fmt.Sprintf("%.6f", ob.Bids[i].Price)
			bidQ = fmt.Sprintf("%.6f", ob.Bids[i].Quantity)
		}
		if i < len(ob.Asks) {
			askP = fmt.Sprintf("%.6f", ob.Asks[i].Price)
			askQ = fmt.Sprintf("%.6f", ob.Asks[i].Quantity)
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", bidP, bidQ, askP, askQ)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}
