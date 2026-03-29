package rendertoon

import (
	"io"

	"github.com/mdnmdn/bits/internal/model"
)

func RenderOrderBook(w io.Writer, res model.Response[model.OrderBook]) error {
	return Render(w, res)
}
