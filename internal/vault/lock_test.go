package vault

import (
	"errors"
	"testing"
)

type fakeLockClient struct {
	store   map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (f *fakeLockClient) ReadSecret(mount, path string) (map[string]interface{}, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	key := mount + "/" + path
	if v, ok := f.store[key]; ok {
		return v, nil
	}
	return nil, nil
}

func (f *fakeLockClient) WriteSecret(mount, path string, data map[string]interface{}) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	key := mount + "/" + path
	f.store[key] = data
	return nil
}

func newFakeLockClient(initial map[string]map[string]interface{}) *fakeLockClient {
	if initial == nil {
		initial = map[string]map[string]interface{}{}
	}
	return &fakeLockClient{store: initial}
}

func TestLockSecret_Success(t *testing.T) {
	client := newFakeLockClient(map[string]map[string]interface{}{
		"secret/myapp": {"API_KEY": "abc123"},
	})
	res := LockSecret(client, "secret", "myapp", "alice")
	if !res.Success {
		t.Fatalf("expected success, got error: %v", res.Error)
	}
	data := client.store["secret/myapp"]
	if _, ok := data[lockMetaKey]; !ok {
		t.Error("expected lock marker to be written")
	}
}

func TestLockSecret_ReadError(t *testing.T) {
	client := newFakeLockClient(nil)
	client.readErr = errors.New("vault unavailable")
	res := LockSecret(client, "secret", "myapp", "alice")
	if res.Success {
		t.Fatal("expected failure")
	}
}

func TestUnlockSecret_Success(t *testing.T) {
	client := newFakeLockClient(map[string]map[string]interface{}{
		"secret/myapp": {"API_KEY": "abc", lockMetaKey: "alice@2024-01-01T00:00:00Z"},
	})
	res := UnlockSecret(client, "secret", "myapp")
	if !res.Success {
		t.Fatalf("expected success, got: %v", res.Error)
	}
	data := client.store["secret/myapp"]
	if _, ok := data[lockMetaKey]; ok {
		t.Error("expected lock marker to be removed")
	}
}

func TestUnlockSecret_NotFound(t *testing.T) {
	client := newFakeLockClient(nil)
	res := UnlockSecret(client, "secret", "missing")
	if res.Success {
		t.Fatal("expected failure for missing secret")
	}
}

func TestLockSecrets_MultipleResults(t *testing.T) {
	client := newFakeLockClient(map[string]map[string]interface{}{
		"secret/a": {"k": "v"},
		"secret/b": {"k": "v"},
	})
	results := LockSecrets(client, "secret", []string{"a", "b"}, "ci-bot")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for path %s, got: %v", r.Path, r.Error)
		}
	}
}
