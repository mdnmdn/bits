package rendertable

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
)

func RenderCandles(w io.Writer, res model.Response[[]model.Candle]) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "OPEN TIME\tOPEN\tHIGH\tLOW\tCLOSE\tVOLUME")
	for _, c := range res.Data {
		vol := "-"
		if c.Volume != nil {
			vol = fmt.Sprintf("%.4f", *c.Volume)
		}
		fmt.Fprintf(tw, "%s\t%.6f\t%.6f\t%.6f\t%.6f\t%s\n",
			c.OpenTime.Format("2006-01-02 15:04"),
			c.Open, c.High, c.Low, c.Close, vol)
	}
	tw.Flush()
	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintf(w, "\n%s\n", note)
	}
	return nil
}
