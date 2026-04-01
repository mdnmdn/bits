package table

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/mdnmdn/bits/model"
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

	if len(info.Symbols) == 1 {
		renderSymbolDetails(w, info.Symbols[0])
		printFooter(w, res)
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "SYMBOL\tORIGINAL\tBASE\tQUOTE\tSTATUS\tMARKET\tPRICE_PREC\tQTY_PREC\tMIN_PRICE\tMAX_PRICE\tMIN_QTY\tMAX_QTY\tSTEPSIZE\tMAKER_FEE\tTAKER_FEE")
	for _, s := range info.Symbols {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			NormalizeSymbol(s.Symbol), s.Symbol, s.BaseAsset, s.QuoteAsset, s.Status, s.Market,
			ptrToStr(s.PricePrecision), ptrToStr(s.QtyPrecision),
			ptrToStr(s.MinPrice), ptrToStr(s.MaxPrice),
			ptrToStr(s.MinQty), ptrToStr(s.MaxQty),
			ptrToStr(s.StepSize), ptrToStr(s.MakerFee), ptrToStr(s.TakerFee))
	}
	_ = tw.Flush()
	printFooter(w, res)
	return nil
}

func renderSymbolDetails(w io.Writer, s model.Symbol) {
	_, _ = fmt.Fprintln(w, "---")
	_, _ = fmt.Fprintf(w, "Symbol        : %s\n", NormalizeSymbol(s.Symbol))
	_, _ = fmt.Fprintf(w, "Original      : %s\n", s.Symbol)
	_, _ = fmt.Fprintf(w, "Base Asset    : %s\n", s.BaseAsset)
	_, _ = fmt.Fprintf(w, "Quote Asset   : %s\n", s.QuoteAsset)
	_, _ = fmt.Fprintf(w, "Status        : %s\n", s.Status)
	_, _ = fmt.Fprintf(w, "Market        : %s\n", s.Market)
	if s.PricePrecision != nil {
		_, _ = fmt.Fprintf(w, "Price Prec    : %d\n", *s.PricePrecision)
	}
	if s.QtyPrecision != nil {
		_, _ = fmt.Fprintf(w, "Qty Precision : %d\n", *s.QtyPrecision)
	}
	if s.MinPrice != nil {
		_, _ = fmt.Fprintf(w, "Min Price     : %v\n", *s.MinPrice)
	}
	if s.MaxPrice != nil {
		_, _ = fmt.Fprintf(w, "Max Price     : %v\n", *s.MaxPrice)
	}
	if s.MinQty != nil {
		_, _ = fmt.Fprintf(w, "Min Qty        : %v\n", *s.MinQty)
	}
	if s.MaxQty != nil {
		_, _ = fmt.Fprintf(w, "Max Qty        : %v\n", *s.MaxQty)
	}
	if s.StepSize != nil {
		_, _ = fmt.Fprintf(w, "Step Size     : %v\n", *s.StepSize)
	}
	if s.MakerFee != nil {
		_, _ = fmt.Fprintf(w, "Maker Fee     : %v\n", *s.MakerFee)
	}
	if s.TakerFee != nil {
		_, _ = fmt.Fprintf(w, "Taker Fee     : %v\n", *s.TakerFee)
	}
	if len(s.Extra) > 0 {
		_, _ = fmt.Fprintln(w, "Extra         :")
		for k, v := range s.Extra {
			_, _ = fmt.Fprintf(w, "  %s: %v\n", k, v)
		}
	}
	_, _ = fmt.Fprintln(w, "---")
}

func ptrToStr[T any](v *T) string {
	if v == nil {
		return "-"
	}
	return fmt.Sprintf("%v", *v)
}
