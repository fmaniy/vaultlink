package vault

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPrintArchiveReportTo_ContainsHeaders(t *testing.T) {
	var buf bytes.Buffer
	printArchiveReportTo(&buf, []ArchiveResult{}, "archive")
	out := buf.String()
	for _, hdr := range []string{"PATH", "STATUS", "DETAIL"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestPrintArchiveReportTo_ShowsSuccess(t *testing.T) {
	at := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	results := []ArchiveResult{
		{Path: "secret/app", ArchivedAt: at},
	}
	var buf bytes.Buffer
	printArchiveReportTo(&buf, results, "archive")
	out := buf.String()
	if !strings.Contains(out, "secret/app") {
		t.Error("expected path in output")
	}
	if !strings.Contains(out, "OK") {
		t.Error("expected OK status in output")
	}
	if !strings.Contains(out, "2024-06-01") {
		t.Error("expected archived timestamp in output")
	}
}

func TestPrintArchiveReportTo_ShowsFailure(t *testing.T) {
	results := []ArchiveResult{
		{Path: "secret/missing", Err: errors.New("secret not found: secret/missing")},
	}
	var buf bytes.Buffer
	printArchiveReportTo(&buf, results, "archive")
	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Error("expected ERROR status in output")
	}
	if !strings.Contains(out, "secret not found") {
		t.Error("expected error message in output")
	}
}

func TestPrintArchiveReportTo_SummaryLine(t *testing.T) {
	at := time.Now().UTC()
	results := []ArchiveResult{
		{Path: "secret/a", ArchivedAt: at},
		{Path: "secret/b", Err: errors.New("some error")},
	}
	var buf bytes.Buffer
	printArchiveReportTo(&buf, results, "unarchive")
	out := buf.String()
	if !strings.Contains(out, "unarchive complete") {
		t.Error("expected summary line with action name")
	}
	if !strings.Contains(out, "1 succeeded") {
		t.Error("expected 1 succeeded in summary")
	}
	if !strings.Contains(out, "1 failed") {
		t.Error("expected 1 failed in summary")
	}
}
