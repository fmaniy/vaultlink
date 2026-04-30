package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintTagReportTo_ContainsHeaders(t *testing.T) {
	var buf strings.Builder
	printTagReportTo(&buf, []TagResult{})
	out := buf.String()
	for _, h := range []string{"PATH", "STATUS", "TAGS"} {
		if !strings.Contains(out, h) {
			t.Errorf("missing header %q in output", h)
		}
	}
}

func TestPrintTagReportTo_ShowsSuccess(t *testing.T) {
	var buf strings.Builder
	printTagReportTo(&buf, []TagResult{
		{Path: "myapp/db", Tags: []string{"prod", "critical"}, Success: true},
	})
	out := buf.String()
	if !strings.Contains(out, "myapp/db") {
		t.Error("expected path in output")
	}
	if !strings.Contains(out, "tagged") {
		t.Error("expected status 'tagged' in output")
	}
	if !strings.Contains(out, "prod") {
		t.Error("expected tag 'prod' in output")
	}
}

func TestPrintTagReportTo_ShowsFailure(t *testing.T) {
	var buf strings.Builder
	printTagReportTo(&buf, []TagResult{
		{Path: "myapp/db", Success: false, Err: errors.New("permission denied")},
	})
	out := buf.String()
	if !strings.Contains(out, "error") {
		t.Error("expected status 'error' in output")
	}
	if !strings.Contains(out, "permission denied") {
		t.Error("expected error message in output")
	}
}

func TestPrintTagReportTo_SummaryLine(t *testing.T) {
	var buf strings.Builder
	printTagReportTo(&buf, []TagResult{
		{Path: "a", Tags: []string{"x"}, Success: true},
		{Path: "b", Success: false, Err: errors.New("fail")},
	})
	out := buf.String()
	if !strings.Contains(out, "Total: 2") {
		t.Error("expected Total: 2 in summary")
	}
	if !strings.Contains(out, "OK: 1") {
		t.Error("expected OK: 1 in summary")
	}
	if !strings.Contains(out, "Failed: 1") {
		t.Error("expected Failed: 1 in summary")
	}
}
