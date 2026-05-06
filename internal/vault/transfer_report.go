package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintTransferReport writes a formatted transfer report to stdout.
func PrintTransferReport(results []TransferResult) {
	printTransferReportTo(os.Stdout, results)
}

func printTransferReportTo(w io.Writer, results []TransferResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSTATUS")
	fmt.Fprintln(tw, "----\t------")

	ok, skipped, failed := 0, 0, 0
	for _, r := range results {
		switch {
		case r.Error != nil:
			fmt.Fprintf(tw, "%s\tERROR: %v\n", r.Path, r.Error)
			failed++
		case r.Skipped:
			fmt.Fprintf(tw, "%s\tSKIPPED\n", r.Path)
			skipped++
		default:
			fmt.Fprintf(tw, "%s\tOK\n", r.Path)
			ok++
		}
	}

	tw.Flush()
	fmt.Fprintf(w, "\nSummary: %d transferred, %d skipped, %d failed\n", ok, skipped, failed)
}

// TransferReportHasFailures returns true if any result contains an error.
func TransferReportHasFailures(results []TransferResult) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}
