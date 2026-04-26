package audit

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintReportTo_ContainsHeader(t *testing.T) {
	r := Report{
		SourceEnv: "staging",
		TargetEnv: "production",
		Diffs:     []DiffResult{},
	}

	var buf bytes.Buffer
	printReportTo(&buf, r)

	out := buf.String()
	if !strings.Contains(out, "staging") {
		t.Errorf("expected source env in output, got: %s", out)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("expected target env in output, got: %s", out)
	}
	if !strings.Contains(out, "KEY") {
		t.Errorf("expected table header KEY in output, got: %s", out)
	}
}

func TestPrintReportTo_ShowsMissingStatus(t *testing.T) {
	r := Report{
		SourceEnv: "dev",
		TargetEnv: "prod",
		Diffs: []DiffResult{
			{Key: "DB_PASSWORD", Status: "missing"},
		},
	}

	var buf bytes.Buffer
	printReportTo(&buf, r)

	out := buf.String()
	if !strings.Contains(out, "DB_PASSWORD") {
		t.Errorf("expected key DB_PASSWORD in output")
	}
	if !strings.Contains(out, "missing") {
		t.Errorf("expected status 'missing' in output")
	}
}

func TestPrintReportTo_SummaryLine(t *testing.T) {
	r := Report{
		SourceEnv: "dev",
		TargetEnv: "prod",
		Diffs: []DiffResult{
			{Key: "A", Status: "ok"},
			{Key: "B", Status: "missing"},
			{Key: "C", Status: "mismatch"},
		},
	}

	var buf bytes.Buffer
	printReportTo(&buf, r)

	out := buf.String()
	if !strings.Contains(out, "1 ok") {
		t.Errorf("expected '1 ok' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 missing") {
		t.Errorf("expected '1 missing' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 mismatch") {
		t.Errorf("expected '1 mismatch' in summary, got: %s", out)
	}
}

func TestSummarize(t *testing.T) {
	diffs := []DiffResult{
		{Status: "ok"},
		{Status: "ok"},
		{Status: "missing"},
		{Status: "mismatch"},
	}
	ok, missing, mismatch := summarize(diffs)
	if ok != 2 || missing != 1 || mismatch != 1 {
		t.Errorf("unexpected summary counts: ok=%d missing=%d mismatch=%d", ok, missing, mismatch)
	}
}
