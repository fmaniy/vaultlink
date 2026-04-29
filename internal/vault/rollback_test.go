package vault

import (
	"errors"
	"testing"
)

// rollbackFakeLogical reuses the fakeLogical pattern from kv_test.go.
type rollbackFakeLogical struct {
	store  map[string]map[string]interface{}
	readErr  error
	writeErr error
}

func (f *rollbackFakeLogical) Read(path string) (*fakeSecret, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	data, ok := f.store[path]
	if !ok {
		return nil, nil
	}
	return &fakeSecret{data: map[string]interface{}{"data": data}}, nil
}

func (f *rollbackFakeLogical) Write(path string, body map[string]interface{}) (*fakeSecret, error) {
	if f.writeErr != nil {
		return nil, f.writeErr
	}
	payload, _ := body["data"].(map[string]interface{})
	f.store[path] = payload
	return nil, nil
}

func newRollbackClient(store map[string]map[string]interface{}) *fakeVaultClient {
	return clientWithFakeLogical(&rollbackFakeLogical{store: store})
}

func TestTakeSnapshot_Success(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/myapp": {"key": "value"},
	}
	client := newRollbackClient(store)
	snap, err := TakeSnapshot(client, "secret", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Path != "myapp" {
		t.Errorf("expected path myapp, got %s", snap.Path)
	}
	if snap.Data["key"] != "value" {
		t.Errorf("expected key=value in snapshot")
	}
}

func TestRestoreSnapshot_Success(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/myapp": {"key": "old"},
	}
	client := newRollbackClient(store)
	snap := &SecretSnapshot{Path: "myapp", Data: map[string]interface{}{"key": "restored"}}
	if err := RestoreSnapshot(client, "secret", snap); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := store["secret/data/myapp"]["key"]
	if got != "restored" {
		t.Errorf("expected restored, got %v", got)
	}
}

func TestRestoreSnapshot_NilSnapshot(t *testing.T) {
	client := newRollbackClient(map[string]map[string]interface{}{})
	if err := RestoreSnapshot(client, "secret", nil); err == nil {
		t.Fatal("expected error for nil snapshot")
	}
}

func TestRollbackSecrets_RollsBackOnError(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/app": {"token": "original"},
	}
	client := newRollbackClient(store)

	err := RollbackSecrets(client, "secret", []string{"app"}, func() error {
		// Mutate then fail
		store["secret/data/app"] = map[string]interface{}{"token": "mutated"}
		return errors.New("operation failed")
	})

	if err == nil {
		t.Fatal("expected error from RollbackSecrets")
	}
	if store["secret/data/app"]["token"] != "original" {
		t.Errorf("expected rollback to restore original value")
	}
}

func TestRollbackSecrets_NoErrorNoRollback(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/app": {"token": "original"},
	}
	client := newRollbackClient(store)

	err := RollbackSecrets(client, "secret", []string{"app"}, func() error {
		store["secret/data/app"] = map[string]interface{}{"token": "updated"}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store["secret/data/app"]["token"] != "updated" {
		t.Errorf("expected updated value to persist")
	}
}
