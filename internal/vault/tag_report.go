package vault

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintTagReport writes a formatted tag operation report to stdout.
func PrintTagReport(results []TagResult) {
	printTagReportTo(os.Stdout, results)
}

func printTagReportTo(w io.Writer, results []TagResult) {
	fmt.Fprintf(w, "%-40s %-12s %s\n", "PATH", "STATUS", "TAGS")
	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 72))

	ok, failed := 0, 0
	for _, r := range results {
		if r.Success {
			ok++
			fmt.Fprintf(w, "%-40s %-12s %s\n", r.Path, "tagged", strings.Join(r.Tags, ", "))
		} else {
			failed++
			fmt.Fprintf(w, "%-40s %-12s %s\n", r.Path, "error", r.Err.Error())
		}
	}

	fmt.Fprintf(w, "%s\n", strings.Repeat("-", 72))
	fmt.Fprintf(w, "Total: %d  OK: %d  Failed: %d\n", len(results), ok, failed)
}
