package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintPatchReportTo_ContainsHeaders(t *testing.T) {
	var sb strings.Builder
	printPatchReportTo(&sb, []PatchResult{})
	out := sb.String()
	for _, h := range []string{"PATH", "KEY", "STATUS", "OLD", "NEW"} {
		if !strings.Contains(out, h) {
			t.Errorf("expected header %q in output", h)
		}
	}
}

func TestPrintPatchReportTo_ShowsUpdated(t *testing.T) {
	var sb strings.Builder
	results := []PatchResult{
		{Path: "secret/app", Key: "db_pass", OldVal: "old", NewVal: "new"},
	}
	printPatchReportTo(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "updated") {
		t.Error("expected 'updated' in output")
	}
	if !strings.Contains(out, "secret/app") {
		t.Error("expected path in output")
	}
}

func TestPrintPatchReportTo_ShowsSkipped(t *testing.T) {
	var sb strings.Builder
	results := []PatchResult{
		{Path: "secret/app", Key: "missing_key", Skipped: true},
	}
	printPatchReportTo(&sb, results)
	if !strings.Contains(sb.String(), "skipped") {
		t.Error("expected 'skipped' in output")
	}
}

func TestPrintPatchReportTo_ShowsError(t *testing.T) {
	var sb strings.Builder
	results := []PatchResult{
		{Path: "secret/app", Key: "k", Err: errors.New("vault down")},
	}
	printPatchReportTo(&sb, results)
	if !strings.Contains(sb.String(), "error") {
		t.Error("expected 'error' in output")
	}
}

func TestPrintPatchReportTo_SummaryLine(t *testing.T) {
	var sb strings.Builder
	results := []PatchResult{
		{Path: "secret/a", Key: "k", OldVal: "x", NewVal: "y"},
		{Path: "secret/b", Key: "k", Skipped: true},
		{Path: "secret/c", Key: "k", Err: errors.New("fail")},
	}
	printPatchReportTo(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "1 updated") {
		t.Errorf("expected '1 updated' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("expected '1 skipped' in summary, got: %s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got: %s", out)
	}
}
