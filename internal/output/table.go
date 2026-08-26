package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// RenderTable prints a clean tabular format using text/tabwriter.
func RenderTable(w io.Writer, headers []string, rows [][]string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	// Write headers
	headerLine := strings.Join(headers, "\t")
	fmt.Fprintln(tw, strings.ToUpper(headerLine))

	// Write divider
	var dividers []string
	for _, h := range headers {
		dividers = append(dividers, strings.Repeat("-", len(h)))
	}
	fmt.Fprintln(tw, strings.Join(dividers, "\t"))

	// Write rows
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}

	tw.Flush()
}
