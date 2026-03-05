package display

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintTableBasic(t *testing.T) {
	out := captureStdout(func() {
		PrintTable(
			[]string{"Name", "Value"},
			[][]string{
				{"Alice", "100"},
				{"Bob", "200"},
			},
		)
	})

	assert.Contains(t, out, "Name")
	assert.Contains(t, out, "Value")
	assert.Contains(t, out, "Alice")
	assert.Contains(t, out, "Bob")
	assert.Contains(t, out, "───")
}

func TestPrintTableEmpty(t *testing.T) {
	out := captureStdout(func() {
		PrintTable([]string{"A", "B"}, nil)
	})

	assert.Contains(t, out, "No data to display")
}

func TestPrintTableAlignmentWithANSI(t *testing.T) {
	out := captureStdout(func() {
		PrintTable(
			[]string{"Name", "Change"},
			[][]string{
				{"Bitcoin", "\033[32m5.25%\033[0m"},
				{"Ethereum", "-1.00%"},
			},
		)
	})

	lines := strings.Split(out, "\n")
	require.True(t, len(lines) >= 4, "expected at least 4 lines")

	// Data rows should have the same visible width despite ANSI codes
	// Skip separator line (index 1) — it uses multi-byte `─` chars
	headerWidth := VisibleWidth(lines[0])
	row1Width := VisibleWidth(lines[2])
	row2Width := VisibleWidth(lines[3])

	assert.Equal(t, headerWidth, row1Width, "header and ANSI row should align")
	assert.Equal(t, row1Width, row2Width, "ANSI row and plain row should align")
}

func TestPrintTableColumnWidthExpandsToFitContent(t *testing.T) {
	out := captureStdout(func() {
		PrintTable(
			[]string{"X"},
			[][]string{
				{"Short"},
				{"A much longer value here"},
			},
		)
	})

	lines := strings.Split(out, "\n")
	require.True(t, len(lines) >= 2)

	// Separator should be at least as wide as the longest content
	sepLine := lines[1]
	assert.True(t, VisibleWidth(sepLine) >= len("A much longer value here"),
		"separator should expand to fit longest cell")
}
