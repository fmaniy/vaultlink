package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintValidateReportTo_ContainsHeaders(t *testing.T) {
	var buf strings.Builder
	printValidateReportTo(&buf, []ValidationResult{})
	out := buf.String()
	if !strings.Contains(out, "PATH") {
		t.Error("expected PATH header")
	}
	if !strings.Contains(out, "STATUS") {
		t.Error("expected STATUS header")
	}
	if !strings.Contains(out, "DETAIL") {
		t.Error("expected DETAIL header")
	}
}

func TestPrintValidateReportTo_ShowsOk(t *testing.T) {
	var buf strings.Builder
	results := []ValidationResult{
		{Path: "app/db", Status: "ok"},
	}
	printValidateReportTo(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "app/db") {
		t.Error("expected path in output")
	}
	if !strings.Contains(out, "OK") {
		t.Error("expected OK status")
	}
}

func TestPrintValidateReportTo_ShowsMissingKeys(t *testing.T) {
	var buf strings.Builder
	results := []ValidationResult{
		{Path: "app/cfg", Status: "missing_keys", Missing: []string{"port"}, Err: errors.New("missing keys: port")},
	}
	printValidateReportTo(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "MISSING KEYS") {
		t.Error("expected MISSING KEYS status")
	}
	if !strings.Contains(out, "missing keys: port") {
		t.Error("expected error detail")
	}
}

func TestPrintValidateReportTo_SummaryLine(t *testing.T) {
	var buf strings.Builder
	results := []ValidationResult{
		{Path: "a", Status: "ok"},
		{Path: "b", Status: "missing_keys", Err: errors.New("missing keys: x")},
		{Path: "c", Status: "not_found", Err: errors.New("secret not found")},
	}
	printValidateReportTo(&buf, results)
	out := buf.String()
	if !strings.Contains(out, "1 ok") {
		t.Error("expected 1 ok in summary")
	}
	if !strings.Contains(out, "1 missing keys") {
		t.Error("expected 1 missing keys in summary")
	}
	if !strings.Contains(out, "1 errors") {
		t.Error("expected 1 errors/not found in summary")
	}
}
