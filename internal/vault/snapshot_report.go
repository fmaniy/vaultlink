package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintSnapshotReport writes a formatted snapshot report to stdout.
func PrintSnapshotReport(results []SnapshotResult) {
	printSnapshotReportTo(os.Stdout, results)
}

func printSnapshotReportTo(w io.Writer, results []SnapshotResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSTATUS\tDETAIL")
	fmt.Fprintln(tw, "----\t------\t------")

	var ok, failed int
	for _, r := range results {
		status := "OK"
		detail := ""
		if !r.OK {
			status = "ERROR"
			detail = r.Error
			failed++
		} else {
			ok++
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Path, status, detail)
	}
	_ = tw.Flush()

	fmt.Fprintf(w, "\nSnapshot complete: %d captured, %d failed\n", ok, failed)
}
