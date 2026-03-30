package cmd

import (
	"io"

	"github.com/mdnmdn/bits/internal/render"
	renderjson "github.com/mdnmdn/bits/internal/render/json"
	rendermd "github.com/mdnmdn/bits/internal/render/markdown"
	rendertoon "github.com/mdnmdn/bits/internal/render/toon"
	renderyaml "github.com/mdnmdn/bits/internal/render/yaml"
	"github.com/mdnmdn/bits/pkg/model"
)

// renderGeneric handles all non-table output formats generically.
// Returns (true, err) when the format was handled; (false, nil) when the
// caller should fall back to its type-specific table renderer.
func renderGeneric[T any](w io.Writer, format render.Format, res model.Response[T]) (bool, error) {
	switch format {
	case render.FormatJSON:
		return true, renderjson.Render(w, res)
	case render.FormatYAML:
		return true, renderyaml.Render(w, res)
	case render.FormatMarkdown:
		return true, rendermd.Render(w, res)
	case render.FormatToon:
		return true, rendertoon.Render(w, res)
	}
	return false, nil
}
