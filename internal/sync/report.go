package sync

import (
	"fmt"
	"io"
	"text/tabwriter"
)

// PrintReport writes a human-readable summary of sync results to w.
func PrintReport(w io.Writer, results []Result) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PATH\tSOURCE\tDESTINATION\tACTION\tERROR")
	fmt.Fprintln(tw, "----\t------\t-----------\t------\t-----")

	for _, r := range results {
		errStr := ""
		if r.Err != nil {
			errStr = r.Err.Error()
		}
		action := r.Action
		if action == "" && r.Err != nil {
			action = "failed"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.Path, r.Source, r.Dest, action, errStr)
	}
	tw.Flush()
}

// Summary returns counts of created, updated, skipped, and failed results.
func Summary(results []Result) (created, updated, skipped, failed int) {
	for _, r := range results {
		if r.Err != nil {
			failed++
			continue
		}
		switch r.Action {
		case "created":
			created++
		case "updated":
			updated++
		case "skipped":
			skipped++
		}
	}
	return
}
