package vault

import (
	"context"
	"errors"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

// fakeLogical is a minimal stub for the logical client used in KV operations.
type fakeLogical struct {
	readFn  func(path string) (*vaultapi.Secret, error)
	writeFn func(path string, data map[string]interface{}) (*vaultapi.Secret, error)
	listFn  func(path string) (*vaultapi.Secret, error)
}

func (f *fakeLogical) ReadWithContext(_ context.Context, path string) (*vaultapi.Secret, error) {
	return f.readFn(path)
}

func (f *fakeLogical) WriteWithContext(_ context.Context, path string, data map[string]interface{}) (*vaultapi.Secret, error) {
	return f.writeFn(path, data)
}

func (f *fakeLogical) ListWithContext(_ context.Context, path string) (*vaultapi.Secret, error) {
	return f.listFn(path)
}

func clientWithFakeLogical(fl *fakeLogical) *Client {
	return &Client{logical: fl}
}

func TestReadSecret_Success(t *testing.T) {
	fl := &fakeLogical{
		readFn: func(path string) (*vaultapi.Secret, error) {
			if path != "secret/data/myapp" {
				t.Fatalf("unexpected path: %s", path)
			}
			return &vaultapi.Secret{
				Data: map[string]interface{}{
					"data": map[string]interface{}{"API_KEY": "abc123"},
				},
			}, nil
		},
	}
	c := clientWithFakeLogical(fl)
	data, err := c.ReadSecret(context.Background(), "secret", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["API_KEY"] != "abc123" {
		t.Errorf("expected API_KEY=abc123, got %q", data["API_KEY"])
	}
}

func TestReadSecret_NotFound(t *testing.T) {
	fl := &fakeLogical{
		readFn: func(_ string) (*vaultapi.Secret, error) { return nil, nil },
	}
	c := clientWithFakeLogical(fl)
	_, err := c.ReadSecret(context.Background(), "secret", "missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestReadSecret_VaultError(t *testing.T) {
	fl := &fakeLogical{
		readFn: func(_ string) (*vaultapi.Secret, error) {
			return nil, errors.New("permission denied")
		},
	}
	c := clientWithFakeLogical(fl)
	_, err := c.ReadSecret(context.Background(), "secret", "app")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListSecrets_Success(t *testing.T) {
	fl := &fakeLogical{
		listFn: func(_ string) (*vaultapi.Secret, error) {
			return &vaultapi.Secret{
				Data: map[string]interface{}{
					"keys": []interface{}{"app1", "app2"},
				},
			}, nil
		},
	}
	c := clientWithFakeLogical(fl)
	keys, err := c.ListSecrets(context.Background(), "secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}
