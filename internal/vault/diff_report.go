package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// PrintDiffReport writes a human-readable diff report to stdout.
func PrintDiffReport(results []DiffResult) {
	printDiffReportTo(os.Stdout, results)
}

func printDiffReportTo(w io.Writer, results []DiffResult) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tKEY\tSTATUS")
	fmt.Fprintln(tw, "----\t---\t------")

	for _, r := range results {
		status := formatDiffStatus(r.Status)
		fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Path, r.Key, status)
	}
	tw.Flush()

	match, missing, extra, mismatch := 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case DiffStatusMatch:
			match++
		case DiffStatusMissing:
			missing++
		case DiffStatusExtra:
			extra++
		case DiffStatusMismatch:
			mismatch++
		}
	}
	fmt.Fprintf(w, "\nSummary: %d match, %d missing, %d extra, %d mismatch\n",
		match, missing, extra, mismatch)
}

func formatDiffStatus(s DiffStatus) string {
	switch s {
	case DiffStatusMatch:
		return colorGreen + "match" + colorReset
	case DiffStatusMissing:
		return colorRed + "missing" + colorReset
	case DiffStatusExtra:
		return colorYellow + "extra" + colorReset
	case DiffStatusMismatch:
		return colorRed + "mismatch" + colorReset
	default:
		return string(s)
	}
}
