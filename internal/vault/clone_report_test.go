package vault

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestPrintCloneReportTo_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	printCloneReportTo(&buf, []CloneResult{}, "dev", "staging")
	out := buf.String()
	for _, h := range []string{"SOURCE ENV", "DEST ENV", "PATH", "STATUS", "ERROR"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q in output", h)
		}
	}
}

func TestPrintCloneReportTo_ShowsResults(t *testing.T) {
	results := []CloneResult{
		{SourcePath: "app/db", DestPath: "app/db", Status: "cloned"},
		{SourcePath: "app/api", DestPath: "app/api", Status: "skipped"},
		{SourcePath: "app/broken", DestPath: "app/broken", Status: "error", Err: errors.New("vault down")},
	}
	var buf bytes.Buffer
	printCloneReportTo(&buf, results, "dev", "staging")
	out := buf.String()

	if !strings.Contains(out, "app/db") {
		t.Error("expected app/db in output")
	}
	if !strings.Contains(out, "skipped") {
		t.Error("expected skipped status in output")
	}
	if !strings.Contains(out, "vault down") {
		t.Error("expected error message in output")
	}
}

func TestPrintCloneReportTo_SummaryLine(t *testing.T) {
	results := []CloneResult{
		{Status: "cloned"},
		{Status: "cloned"},
		{Status: "skipped"},
		{Status: "error", Err: errors.New("oops")},
	}
	var buf bytes.Buffer
	printCloneReportTo(&buf, results, "dev", "prod")
	out := buf.String()

	if !strings.Contains(out, "2 cloned") {
		t.Errorf("expected '2 cloned' in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("expected '1 skipped' in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "1 error") {
		t.Errorf("expected '1 error' in summary, got:\n%s", out)
	}
}
