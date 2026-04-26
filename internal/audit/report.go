package audit

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// DiffResult holds the result of comparing a single secret key across environments.
type DiffResult struct {
	Key    string
	Source string
	Target string
	Status string // "ok", "missing", "mismatch"
}

// Report contains the full audit output for a source/target pair.
type Report struct {
	SourceEnv string
	TargetEnv string
	Diffs     []DiffResult
}

// PrintReport writes a formatted audit report to stdout.
func PrintReport(r Report) {
	printReportTo(os.Stdout, r)
}

func printReportTo(w io.Writer, r Report) {
	fmt.Fprintf(w, "\nAudit Report: %s → %s\n", r.SourceEnv, r.TargetEnv)
	fmt.Fprintf(w, "%s\n", repeatChar('-', 50))

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw, "KEY\tSTATUS\tDETAIL")

	for _, d := range r.Diffs {
		detail := ""
		switch d.Status {
		case "missing":
			detail = fmt.Sprintf("key not found in %s", r.TargetEnv)
		case "mismatch":
			detail = "values differ between environments"
		case "ok":
			detail = "in sync"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", d.Key, d.Status, detail)
	}
	tw.Flush()

	ok, missing, mismatch := summarize(r.Diffs)
	fmt.Fprintf(w, "\nSummary: %d ok, %d missing, %d mismatch\n", ok, missing, mismatch)
}

func summarize(diffs []DiffResult) (ok, missing, mismatch int) {
	for _, d := range diffs {
		switch d.Status {
		case "ok":
			ok++
		case "missing":
			missing++
		case "mismatch":
			mismatch++
		}
	}
	return
}

func repeatChar(c rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
