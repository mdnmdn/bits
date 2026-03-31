package rendertoon

import (
	"io"

	"github.com/mdnmdn/bits/pkg/model"
)

func RenderPrice(w io.Writer, res model.Response[model.CoinPrice]) error {
	return Render(w, res)
}
