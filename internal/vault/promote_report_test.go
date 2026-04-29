package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintPromoteReportTo_ContainsHeaders(t *testing.T) {
	var buf strings.Builder
	printPromoteReportTo(&buf, []PromoteResult{}, "staging", "production")
	out := buf.String()

	for _, want := range []string{"PATH", "STATUS", "DETAIL", "staging", "production"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q", want)
		}
	}
}

func TestPrintPromoteReportTo_ShowsSuccess(t *testing.T) {
	var buf strings.Builder
	results := []PromoteResult{
		{Path: "app/db", Success: true},
	}
	printPromoteReportTo(&buf, results, "staging", "production")
	out := buf.String()

	if !strings.Contains(out, "app/db") {
		t.Error("expected path in output")
	}
	if !strings.Contains(out, "promoted") {
		t.Error("expected 'promoted' status in output")
	}
}

func TestPrintPromoteReportTo_ShowsFailure(t *testing.T) {
	var buf strings.Builder
	results := []PromoteResult{
		{Path: "app/api", Success: false, Error: errors.New("already exists")},
	}
	printPromoteReportTo(&buf, results, "staging", "production")
	out := buf.String()

	if !strings.Contains(out, "failed") {
		t.Error("expected 'failed' status in output")
	}
	if !strings.Contains(out, "already exists") {
		t.Error("expected error detail in output")
	}
}

func TestPrintPromoteReportTo_SummaryLine(t *testing.T) {
	var buf strings.Builder
	results := []PromoteResult{
		{Path: "a", Success: true},
		{Path: "b", Success: false, Error: errors.New("err")},
	}
	printPromoteReportTo(&buf, results, "dev", "prod")
	out := buf.String()

	if !strings.Contains(out, "1 promoted") {
		t.Errorf("expected '1 promoted' in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got:\n%s", out)
	}
}
