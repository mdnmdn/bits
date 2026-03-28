package rendertoon

import (
	"bytes"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/mdnmdn/bits/internal/model"
	"github.com/mdnmdn/bits/internal/render"
	"gopkg.in/yaml.v3"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			PaddingLeft(1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	noteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true).
			PaddingLeft(1)
)

// Render writes res as a lipgloss-styled terminal document to w.
// It uses a rounded box containing the data serialised as YAML.
func Render[T any](w io.Writer, res model.Response[T]) error {
	header := headerStyle.Render("◈ " + render.ProviderLabel(res))

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(res.Data); err != nil {
		return err
	}
	// trim trailing newline so the box doesn't get extra padding
	content := buf.String()
	if len(content) > 0 && content[len(content)-1] == '\n' {
		content = content[:len(content)-1]
	}

	fmt.Fprintln(w, header)
	fmt.Fprintln(w, boxStyle.Render(content))

	if note := render.FallbackFootnote(res); note != "" {
		fmt.Fprintln(w, noteStyle.Render("† "+note))
	}
	return nil
}
