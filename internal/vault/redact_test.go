package vault

import (
	"errors"
	"testing"
)

func newRedactClient(data map[string]interface{}, writeErr error) *fakeReadWriter {
	return &fakeReadWriter{data: data, writeErr: writeErr}
}

func TestRedactSecret_RedactsMatchingKeys(t *testing.T) {
	data := map[string]interface{}{"password": "s3cr3t", "host": "localhost"}
	c := newRedactClient(data, nil)

	results := RedactSecrets(c, "secret", []string{"app/cfg"}, []string{"password"}, "***", false)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Err != nil {
		t.Fatalf("unexpected error: %v", r.Err)
	}
	if r.Redacted != 1 {
		t.Errorf("expected 1 redacted, got %d", r.Redacted)
	}
	if r.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", r.Skipped)
	}
	if c.written["password"] != "***" {
		t.Errorf("expected password to be redacted, got %v", c.written["password"])
	}
	if c.written["host"] != "localhost" {
		t.Errorf("expected host to be unchanged, got %v", c.written["host"])
	}
}

func TestRedactSecret_DryRunDoesNotWrite(t *testing.T) {
	data := map[string]interface{}{"token": "abc123", "name": "app"}
	c := newRedactClient(data, nil)

	RedactSecrets(c, "secret", []string{"app/cfg"}, []string{"token"}, "***", true)

	if c.writeCalled {
		t.Error("expected no write in dry-run mode")
	}
}

func TestRedactSecret_NoMatchingKeys(t *testing.T) {
	data := map[string]interface{}{"host": "localhost", "port": "5432"}
	c := newRedactClient(data, nil)

	results := RedactSecrets(c, "secret", []string{"app/db"}, []string{"password"}, "***", false)

	r := results[0]
	if r.Redacted != 0 {
		t.Errorf("expected 0 redacted, got %d", r.Redacted)
	}
	if c.writeCalled {
		t.Error("expected no write when no keys match")
	}
}

func TestRedactSecret_ReadError(t *testing.T) {
	c := &fakeReadWriter{readErr: errors.New("permission denied")}

	results := RedactSecrets(c, "secret", []string{"app/cfg"}, []string{"password"}, "***", false)

	if results[0].Err == nil {
		t.Error("expected error on read failure")
	}
}

func TestMatchesPattern_CaseInsensitive(t *testing.T) {
	if !matchesPattern("DB_PASSWORD", []string{"password"}) {
		t.Error("expected case-insensitive match for DB_PASSWORD")
	}
	if matchesPattern("host", []string{"password", "token"}) {
		t.Error("expected no match for host")
	}
}
