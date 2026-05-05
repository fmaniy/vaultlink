package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintAnnotateReportTo_ContainsHeaders(t *testing.T) {
	var buf strings.Builder
	printAnnotateReportTo(&buf, []AnnotateResult{})
	out := buf.String()

	for _, hdr := range []string{"PATH", "KEY", "VALUE", "STATUS"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestPrintAnnotateReportTo_ShowsSuccess(t *testing.T) {
	var buf strings.Builder
	results := []AnnotateResult{
		{Path: "secret/data/app", Key: "owner", Value: "team-a", Updated: true},
	}
	printAnnotateReportTo(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "secret/data/app") {
		t.Error("expected path in output")
	}
	if !strings.Contains(out, "updated") {
		t.Error("expected 'updated' status")
	}
}

func TestPrintAnnotateReportTo_ShowsFailure(t *testing.T) {
	var buf strings.Builder
	results := []AnnotateResult{
		{Path: "secret/data/missing", Key: "env", Value: "prod", Err: errors.New("not found")},
	}
	printAnnotateReportTo(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "error") {
		t.Error("expected error in output")
	}
	if !strings.Contains(out, "not found") {
		t.Error("expected error message in output")
	}
}

func TestPrintAnnotateReportTo_SummaryLine(t *testing.T) {
	var buf strings.Builder
	results := []AnnotateResult{
		{Path: "secret/data/a", Key: "k", Value: "v", Updated: true},
		{Path: "secret/data/b", Key: "k", Value: "v", Err: errors.New("boom")},
	}
	printAnnotateReportTo(&buf, results)
	out := buf.String()

	if !strings.Contains(out, "1 updated") {
		t.Errorf("expected '1 updated' in summary, got:\n%s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected '1 failed' in summary, got:\n%s", out)
	}
}
