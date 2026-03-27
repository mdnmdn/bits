package rendertable

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
)

func RenderServerTime(w io.Writer, res model.Response[model.ServerTime]) error {
	st := res.Data
	fmt.Fprintf(w, "Server Time : %s\n", st.Time.Format("2006-01-02 15:04:05 UTC"))
	if st.LocalTime != nil {
		fmt.Fprintf(w, "Local Time  : %s\n", st.LocalTime.Format("2006-01-02 15:04:05 UTC"))
	}
	if st.Latency != nil {
		fmt.Fprintf(w, "Latency     : %s\n", st.Latency)
	}
	if st.ClockSkew != nil {
		fmt.Fprintf(w, "Clock Skew  : %s\n", st.ClockSkew)
	}
	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintf(w, "\n%s\n", note)
	}
	return nil
}
