package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintArchiveReport writes a formatted archive/unarchive report to stdout.
func PrintArchiveReport(results []ArchiveResult, action string) {
	printArchiveReportTo(os.Stdout, results, action)
}

func printArchiveReportTo(w io.Writer, results []ArchiveResult, action string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "PATH\tSTATUS\tDETAIL\n")
	fmt.Fprintf(tw, "----\t------\t------\n")

	var succeeded, failed int
	for _, r := range results {
		if r.Err != nil {
			fmt.Fprintf(tw, "%s\tERROR\t%s\n", r.Path, r.Err.Error())
			failed++
		} else {
			detail := ""
			if !r.ArchivedAt.IsZero() {
				detail = r.ArchivedAt.Format("2006-01-02T15:04:05Z")
			}
			fmt.Fprintf(tw, "%s\tOK\t%s\n", r.Path, detail)
			succeeded++
		}
	}

	tw.Flush()
	fmt.Fprintf(w, "\n%s complete: %d succeeded, %d failed\n", action, succeeded, failed)
}
