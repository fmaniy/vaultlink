package sync_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/vaultlink/internal/sync"
)

func TestPrintReport_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	sync.PrintReport(&buf, nil)
	out := buf.String()
	if !strings.Contains(out, "PATH") || !strings.Contains(out, "ACTION") {
		t.Errorf("report missing expected headers, got:\n%s", out)
	}
}

func TestPrintReport_ShowsResults(t *testing.T) {
	results := []sync.Result{
		{Path: "app/config", Source: "staging", Dest: "production", Action: "created"},
		{Path: "app/db", Source: "staging", Dest: "production", Action: "skipped"},
	}
	var buf bytes.Buffer
	sync.PrintReport(&buf, results)
	out := buf.String()
	for _, want := range []string{"app/config", "created", "app/db", "skipped"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in report output:\n%s", want, out)
		}
	}
}

func TestSummary(t *testing.T) {
	results := []sync.Result{
		{Action: "created"},
		{Action: "created"},
		{Action: "updated"},
		{Action: "skipped"},
		{Err: fmt.Errorf("boom")},
	}
	created, updated, skipped, failed := sync.Summary(results)
	if created != 2 || updated != 1 || skipped != 1 || failed != 1 {
		t.Errorf("unexpected summary: created=%d updated=%d skipped=%d failed=%d",
			created, updated, skipped, failed)
	}
}
