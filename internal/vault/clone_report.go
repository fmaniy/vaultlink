package vault

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// PrintCloneReport writes a formatted clone results table to stdout.
func PrintCloneReport(results []CloneResult, srcEnv, dstEnv string) {
	printCloneReportTo(os.Stdout, results, srcEnv, dstEnv)
}

func printCloneReportTo(w io.Writer, results []CloneResult, srcEnv, dstEnv string) {
	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "SOURCE ENV\tDEST ENV\tPATH\tSTATUS\tERROR\n")
	fmt.Fprintf(tw, "----------\t--------\t----\t------\t-----\n")

	var cloned, skipped, errored int
	for _, r := range results {
		errMsg := ""
		if r.Err != nil {
			errMsg = r.Err.Error()
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", srcEnv, dstEnv, r.SourcePath, r.Status, errMsg)
		switch r.Status {
		case "cloned":
			cloned++
		case "skipped":
			skipped++
		case "error":
			errored++
		}
	}

	tw.Flush()
	fmt.Fprintf(w, "\nSummary: %d cloned, %d skipped, %d error(s)\n", cloned, skipped, errored)
}
