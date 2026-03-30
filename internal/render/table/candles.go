package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderCandles(w io.Writer, res model.Response[[]model.Candle]) error {
	printHeader(w, res)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "OPEN TIME\tOPEN\tHIGH\tLOW\tCLOSE\tVOLUME")
	for _, c := range res.Data {
		vol := "-"
		if c.Volume != nil {
			vol = fmt.Sprintf("%.4f", *c.Volume)
		}
		_, _ = fmt.Fprintf(tw, "%s\t%.6f\t%.6f\t%.6f\t%.6f\t%s\n",
			c.OpenTime.Format("2006-01-02 15:04"),
			c.Open, c.High, c.Low, c.Close, vol)
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}
