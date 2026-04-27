package vault

import (
	"errors"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

// fakeListLogical reuses the fakeLogical helper already established in
// kv_test.go (same package), adding list support.
type fakeListLogical struct {
	listData map[string]interface{}
	listErr  error
}

func (f *fakeListLogical) Read(path string) (*vaultapi.Secret, error) {
	return nil, nil
}
func (f *fakeListLogical) Write(path string, data map[string]interface{}) (*vaultapi.Secret, error) {
	return nil, nil
}
func (f *fakeListLogical) List(path string) (*vaultapi.Secret, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listData == nil {
		return nil, nil
	}
	return &vaultapi.Secret{Data: f.listData}, nil
}

func listClient(fl *fakeListLogical) *Client {
	return &Client{logical: fl}
}

func TestListSecrets_Success(t *testing.T) {
	fl := &fakeListLogical{
		listData: map[string]interface{}{
			"keys": []interface{}{"alpha", "beta", "gamma/"},
		},
	}
	c := listClient(fl)

	keys, err := c.ListSecrets("secret", "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	// trailing slash should be stripped
	if keys[2] != "gamma" {
		t.Errorf("expected \"gamma\", got %q", keys[2])
	}
}

func TestListSecrets_EmptyPath(t *testing.T) {
	fl := &fakeListLogical{
		listData: map[string]interface{}{
			"keys": []interface{}{"svcA", "svcB"},
		},
	}
	c := listClient(fl)

	keys, err := c.ListSecrets("secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestListSecrets_NilResponse(t *testing.T) {
	fl := &fakeListLogical{}
	c := listClient(fl)

	keys, err := c.ListSecrets("secret", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}

func TestListSecrets_VaultError(t *testing.T) {
	fl := &fakeListLogical{listErr: errors.New("permission denied")}
	c := listClient(fl)

	_, err := c.ListSecrets("secret", "myapp")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildMetaPath(t *testing.T) {
	cases := []struct {
		mount, path, want string
	}{
		{"secret", "myapp", "secret/metadata/myapp"},
		{"secret", "", "secret/metadata"},
		{"secret", "/myapp/", "secret/metadata/myapp"},
	}
	for _, tc := range cases {
		got := buildMetaPath(tc.mount, tc.path)
		if got != tc.want {
			t.Errorf("buildMetaPath(%q, %q) = %q, want %q", tc.mount, tc.path, got, tc.want)
		}
	}
}
