package vault

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintDiffReportTo_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	printDiffReportTo(&buf, []DiffResult{})
	out := buf.String()
	if !strings.Contains(out, "PATH") || !strings.Contains(out, "KEY") || !strings.Contains(out, "STATUS") {
		t.Errorf("expected headers in output, got: %s", out)
	}
}

func TestPrintDiffReportTo_ShowsResults(t *testing.T) {
	results := []DiffResult{
		{Path: "secret/app", Key: "DB_PASS", Status: DiffStatusMissing},
		{Path: "secret/app", Key: "API_KEY", Status: DiffStatusMatch},
	}
	var buf bytes.Buffer
	printDiffReportTo(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "secret/app") {
		t.Errorf("expected path in output, got: %s", out)
	}
	if !strings.Contains(out, "DB_PASS") {
		t.Errorf("expected key DB_PASS in output, got: %s", out)
	}
	if !strings.Contains(out, "API_KEY") {
		t.Errorf("expected key API_KEY in output, got: %s", out)
	}
}

func TestPrintDiffReportTo_SummaryLine(t *testing.T) {
	results := []DiffResult{
		{Path: "secret/app", Key: "A", Status: DiffStatusMatch},
		{Path: "secret/app", Key: "B", Status: DiffStatusMissing},
		{Path: "secret/app", Key: "C", Status: DiffStatusExtra},
		{Path: "secret/app", Key: "D", Status: DiffStatusMismatch},
	}
	var buf bytes.Buffer
	printDiffReportTo(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "Summary:") {
		t.Errorf("expected Summary line in output, got: %s", out)
	}
	if !strings.Contains(out, "1 match") {
		t.Errorf("expected '1 match' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 missing") {
		t.Errorf("expected '1 missing' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 extra") {
		t.Errorf("expected '1 extra' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 mismatch") {
		t.Errorf("expected '1 mismatch' in summary, got: %s", out)
	}
}
