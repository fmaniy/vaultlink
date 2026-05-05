package vault

import (
	"errors"
	"testing"
)

func newSanitizeClient(data map[string]interface{}, writeErr error) *Client {
	fl := &fakeLogical{
		readData: map[string]interface{}{"data": data},
		writeErr: writeErr,
	}
	return clientWithFakeLogical(fl)
}

func TestSanitizeSecret_RemovesByPrefix(t *testing.T) {
	c := newSanitizeClient(map[string]interface{}{
		"debug_token": "abc",
		"api_key":     "xyz",
	}, nil)

	results := SanitizeSecrets(c, []string{"secret/data/app"}, []string{"debug_"}, nil, false)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Status != "sanitized" {
		t.Errorf("expected sanitized, got %s", r.Status)
	}
	if len(r.Removed) != 1 || r.Removed[0] != "debug_token" {
		t.Errorf("unexpected removed keys: %v", r.Removed)
	}
}

func TestSanitizeSecret_RemovesBySuffix(t *testing.T) {
	c := newSanitizeClient(map[string]interface{}{
		"password_tmp": "secret",
		"username":     "admin",
	}, nil)

	results := SanitizeSecrets(c, []string{"secret/data/app"}, nil, []string{"_tmp"}, false)
	r := results[0]
	if r.Status != "sanitized" {
		t.Errorf("expected sanitized, got %s", r.Status)
	}
	if len(r.Removed) != 1 || r.Removed[0] != "password_tmp" {
		t.Errorf("unexpected removed keys: %v", r.Removed)
	}
}

func TestSanitizeSecret_CleanWhenNoMatch(t *testing.T) {
	c := newSanitizeClient(map[string]interface{}{
		"api_key": "xyz",
	}, nil)

	results := SanitizeSecrets(c, []string{"secret/data/app"}, []string{"debug_"}, nil, false)
	r := results[0]
	if r.Status != "clean" {
		t.Errorf("expected clean, got %s", r.Status)
	}
	if len(r.Removed) != 0 {
		t.Errorf("expected no removed keys, got %v", r.Removed)
	}
}

func TestSanitizeSecret_DryRunDoesNotWrite(t *testing.T) {
	fl := &fakeLogical{
		readData: map[string]interface{}{"data": map[string]interface{}{"debug_x": "1", "keep": "2"}},
	}
	c := clientWithFakeLogical(fl)

	results := SanitizeSecrets(c, []string{"secret/data/app"}, []string{"debug_"}, nil, true)
	r := results[0]
	if r.Status != "dry-run" {
		t.Errorf("expected dry-run, got %s", r.Status)
	}
	if fl.writeCount != 0 {
		t.Errorf("expected no writes in dry-run, got %d", fl.writeCount)
	}
}

func TestSanitizeSecret_ReadError(t *testing.T) {
	fl := &fakeLogical{readErr: errors.New("vault unavailable")}
	c := clientWithFakeLogical(fl)

	results := SanitizeSecrets(c, []string{"secret/data/app"}, []string{"debug_"}, nil, false)
	r := results[0]
	if r.Status != "error" {
		t.Errorf("expected error, got %s", r.Status)
	}
	if r.Err == nil {
		t.Error("expected non-nil error")
	}
}
