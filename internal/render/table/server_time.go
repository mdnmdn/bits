package rendertable

import (
	"fmt"
	"io"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderServerTime(w io.Writer, res model.Response[model.ServerTime]) error {
	printHeader(w, res)
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
	printFooter(w, res)
	return nil
}
