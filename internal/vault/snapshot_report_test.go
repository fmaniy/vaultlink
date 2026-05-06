package vault

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintSnapshotReportTo_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	printSnapshotReportTo(&buf, []SnapshotResult{})
	out := buf.String()
	for _, h := range []string{"PATH", "STATUS", "DETAIL"} {
		if !strings.Contains(out, h) {
			t.Errorf("expected header %q in output", h)
		}
	}
}

func TestPrintSnapshotReportTo_ShowsResults(t *testing.T) {
	results := []SnapshotResult{
		{Path: "secret/db", OK: true},
		{Path: "secret/api", OK: false, Error: "permission denied"},
	}
	var buf bytes.Buffer
	printSnapshotReportTo(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "secret/db") {
		t.Error("expected secret/db in output")
	}
	if !strings.Contains(out, "ERROR") {
		t.Error("expected ERROR status in output")
	}
	if !strings.Contains(out, "permission denied") {
		t.Error("expected error detail in output")
	}
}

func TestPrintSnapshotReportTo_SummaryLine(t *testing.T) {
	results := []SnapshotResult{
		{Path: "a", OK: true},
		{Path: "b", OK: true},
		{Path: "c", OK: false, Error: "err"},
	}
	var buf bytes.Buffer
	printSnapshotReportTo(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "2 captured") {
		t.Errorf("expected '2 captured' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got: %s", out)
	}
}
