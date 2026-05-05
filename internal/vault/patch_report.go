package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintPatchReport writes a formatted patch report to stdout.
func PrintPatchReport(results []PatchResult) {
	printPatchReportTo(os.Stdout, results)
}

func printPatchReportTo(w io.Writer, results []PatchResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tKEY\tSTATUS\tOLD\tNEW")
	fmt.Fprintln(tw, "----\t---\t------\t---\t---")

	var ok, skipped, failed int
	for _, r := range results {
		status := "updated"
		switch {
		case r.Err != nil:
			status = "error"
			failed++
		case r.Skipped:
			status = "skipped"
			skipped++
		default:
			ok++
		}
		errStr := ""
		if r.Err != nil {
			errStr = r.Err.Error()
		}
		oldVal := r.OldVal
		if errStr != "" {
			oldVal = errStr
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.Path, r.Key, status, oldVal, r.NewVal)
	}
	tw.Flush()
	fmt.Fprintf(w, "\nSummary: %d updated, %d skipped, %d failed\n", ok, skipped, failed)
}
