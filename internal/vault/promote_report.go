package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintPromoteReport writes a formatted promotion report to stdout.
func PrintPromoteReport(results []PromoteResult, src, dst string) {
	printPromoteReportTo(os.Stdout, results, src, dst)
}

func printPromoteReportTo(w io.Writer, results []PromoteResult, src, dst string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "Promoting: %s → %s\n", src, dst)
	fmt.Fprintln(tw, "PATH\tSTATUS\tDETAIL")
	fmt.Fprintln(tw, "----\t------\t------")

	succeeded, failed := 0, 0
	for _, r := range results {
		if r.Success {
			fmt.Fprintf(tw, "%s\t✓ promoted\t\n", r.Path)
			succeeded++
		} else {
			fmt.Fprintf(tw, "%s\t✗ failed\t%s\n", r.Path, r.Error)
			failed++
		}
	}

	fmt.Fprintln(tw, "")
	fmt.Fprintf(tw, "Summary: %d promoted, %d failed\n", succeeded, failed)
}
