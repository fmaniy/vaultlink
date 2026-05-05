package vault

import (
	"errors"
	"testing"
)

type fakeProtectClient struct {
	store   map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (f *fakeProtectClient) ReadSecret(_, path string) (map[string]interface{}, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	v, ok := f.store[path]
	if !ok {
		return nil, nil
	}
	copy := make(map[string]interface{}, len(v))
	for k, val := range v {
		copy[k] = val
	}
	return copy, nil
}

func (f *fakeProtectClient) WriteSecret(_, path string, data map[string]interface{}) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.store[path] = data
	return nil
}

func newProtectClient(paths ...string) *fakeProtectClient {
	s := make(map[string]map[string]interface{})
	for _, p := range paths {
		s[p] = map[string]interface{}{"key": "value"}
	}
	return &fakeProtectClient{store: s}
}

func TestProtectSecret_Success(t *testing.T) {
	c := newProtectClient("secret/foo")
	r := ProtectSecret(c, "secret", "secret/foo")
	if r.Error != nil {
		t.Fatalf("unexpected error: %v", r.Error)
	}
	if !r.Protected || r.Skipped {
		t.Errorf("expected Protected=true Skipped=false, got %+v", r)
	}
	if c.store["secret/foo"][protectMetaKey] != "true" {
		t.Error("expected __protected flag to be set")
	}
}

func TestProtectSecret_AlreadyProtected(t *testing.T) {
	c := newProtectClient("secret/foo")
	c.store["secret/foo"][protectMetaKey] = "true"
	r := ProtectSecret(c, "secret", "secret/foo")
	if !r.Skipped {
		t.Error("expected Skipped=true for already-protected secret")
	}
}

func TestProtectSecret_NotFound(t *testing.T) {
	c := newProtectClient()
	r := ProtectSecret(c, "secret", "secret/missing")
	if r.Error == nil {
		t.Error("expected error for missing secret")
	}
}

func TestProtectSecret_ReadError(t *testing.T) {
	c := &fakeProtectClient{store: map[string]map[string]interface{}{}, readErr: errors.New("vault down")}
	r := ProtectSecret(c, "secret", "secret/foo")
	if r.Error == nil {
		t.Error("expected read error to be propagated")
	}
}

func TestUnprotectSecret_Success(t *testing.T) {
	c := newProtectClient("secret/foo")
	c.store["secret/foo"][protectMetaKey] = "true"
	r := UnprotectSecret(c, "secret", "secret/foo")
	if r.Error != nil {
		t.Fatalf("unexpected error: %v", r.Error)
	}
	if r.Protected || r.Skipped {
		t.Errorf("expected Protected=false Skipped=false, got %+v", r)
	}
	if _, exists := c.store["secret/foo"][protectMetaKey]; exists {
		t.Error("expected __protected flag to be removed")
	}
}

func TestUnprotectSecret_NotProtected(t *testing.T) {
	c := newProtectClient("secret/foo")
	r := UnprotectSecret(c, "secret", "secret/foo")
	if !r.Skipped {
		t.Error("expected Skipped=true for non-protected secret")
	}
}

func TestProtectSecrets_Multiple(t *testing.T) {
	c := newProtectClient("a", "b", "c")
	results := ProtectSecrets(c, "secret", []string{"a", "b", "c"}, false)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
	}
}
