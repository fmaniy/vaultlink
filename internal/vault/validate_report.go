package vault

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PrintValidateReport writes a formatted validation report to stdout.
func PrintValidateReport(results []ValidationResult) {
	printValidateReportTo(os.Stdout, results)
}

func printValidateReportTo(w io.Writer, results []ValidationResult) {
	fmt.Fprintf(w, "%-50s %-14s %s\n", "PATH", "STATUS", "DETAIL")
	fmt.Fprintln(w, strings.Repeat("-", 90))

	ok, warn, fail := 0, 0, 0
	for _, r := range results {
		detail := ""
		if r.Err != nil {
			detail = r.Err.Error()
		}
		fmt.Fprintf(w, "%-50s %-14s %s\n", r.Path, formatValidateStatus(r.Status), detail)
		switch r.Status {
		case "ok":
			ok++
		case "missing_keys":
			warn++
		default:
			fail++
		}
	}

	fmt.Fprintln(w, strings.Repeat("-", 90))
	fmt.Fprintf(w, "Summary: %d ok, %d missing keys, %d errors/not found\n", ok, warn, fail)
}

func formatValidateStatus(status string) string {
	switch status {
	case "ok":
		return "OK"
	case "missing_keys":
		return "MISSING KEYS"
	case "not_found":
		return "NOT FOUND"
	case "error":
		return "ERROR"
	default:
		return strings.ToUpper(status)
	}
}
