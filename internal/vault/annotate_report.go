package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintAnnotateReport writes a formatted annotation report to stdout.
func PrintAnnotateReport(results []AnnotateResult) {
	printAnnotateReportTo(os.Stdout, results)
}

func printAnnotateReportTo(w io.Writer, results []AnnotateResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tKEY\tVALUE\tSTATUS")
	fmt.Fprintln(tw, "----\t---\t-----\t------")

	ok, failed := 0, 0
	for _, r := range results {
		status := "updated"
		if r.Err != nil {
			status = fmt.Sprintf("error: %v", r.Err)
			failed++
		} else {
			ok++
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Path, r.Key, r.Value, status)
	}

	tw.Flush()
	fmt.Fprintf(w, "\nSummary: %d updated, %d failed\n", ok, failed)
}
