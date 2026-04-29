package vault

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintSearchReport writes a formatted search results table to stdout.
func PrintSearchReport(query string, results []SearchResult) {
	printSearchReportTo(os.Stdout, query, results)
}

func printSearchReportTo(w io.Writer, query string, results []SearchResult) {
	fmt.Fprintf(w, "Search results for %q:\n", query)
	fmt.Fprintf(w, "%-50s  %s\n", "PATH", "MATCHED KEYS")
	fmt.Fprintln(w, strings.Repeat("-", 80))

	if len(results) == 0 {
		fmt.Fprintln(w, "No matches found.")
		return
	}

	for _, r := range results {
		fmt.Fprintf(w, "%-50s  %s\n", r.Path, strings.Join(r.MatchedKeys, ", "))
	}

	fmt.Fprintln(w, strings.Repeat("-", 80))
	fmt.Fprintf(w, "Total: %d secret(s) matched.\n", len(results))
}
