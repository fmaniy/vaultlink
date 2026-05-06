package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestPrintTransferReportTo_ContainsHeaders(t *testing.T) {
	var sb strings.Builder
	printTransferReportTo(&sb, []TransferResult{})
	out := sb.String()
	if !strings.Contains(out, "PATH") {
		t.Error("expected PATH header")
	}
	if !strings.Contains(out, "STATUS") {
		t.Error("expected STATUS header")
	}
}

func TestPrintTransferReportTo_ShowsOK(t *testing.T) {
	results := []TransferResult{
		{Path: "secret/data/app/db"},
	}
	var sb strings.Builder
	printTransferReportTo(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "OK") {
		t.Error("expected OK status")
	}
	if !strings.Contains(out, "secret/data/app/db") {
		t.Error("expected path in output")
	}
}

func TestPrintTransferReportTo_ShowsSkipped(t *testing.T) {
	results := []TransferResult{
		{Path: "secret/data/app/db", Skipped: true},
	}
	var sb strings.Builder
	printTransferReportTo(&sb, results)
	if !strings.Contains(sb.String(), "SKIPPED") {
		t.Error("expected SKIPPED status")
	}
}

func TestPrintTransferReportTo_ShowsError(t *testing.T) {
	results := []TransferResult{
		{Path: "secret/data/app/db", Error: errors.New("vault down")},
	}
	var sb strings.Builder
	printTransferReportTo(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "ERROR") {
		t.Error("expected ERROR status")
	}
	if !strings.Contains(out, "vault down") {
		t.Error("expected error message")
	}
}

func TestPrintTransferReportTo_SummaryLine(t *testing.T) {
	results := []TransferResult{
		{Path: "a"},
		{Path: "b", Skipped: true},
		{Path: "c", Error: errors.New("boom")},
	}
	var sb strings.Builder
	printTransferReportTo(&sb, results)
	out := sb.String()
	if !strings.Contains(out, "1 transferred") {
		t.Errorf("expected 1 transferred, got: %s", out)
	}
	if !strings.Contains(out, "1 skipped") {
		t.Errorf("expected 1 skipped, got: %s", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("expected 1 failed, got: %s", out)
	}
}
