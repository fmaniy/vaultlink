package vault

import (
	"context"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

// fakeLogicalForMounts satisfies the logicalClient interface for mount tests.
type fakeLogicalForMounts struct {
	data map[string]interface{}
	err  error
}

func (f *fakeLogicalForMounts) Read(_ string) (*vaultapi.Secret, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &vaultapi.Secret{Data: f.data}, nil
}

func (f *fakeLogicalForMounts) Write(_ string, _ map[string]interface{}) (*vaultapi.Secret, error) {
	return nil, nil
}

func (f *fakeLogicalForMounts) List(_ string) (*vaultapi.Secret, error) {
	return nil, nil
}

func mountsData() map[string]interface{} {
	return map[string]interface{}{
		"secret/": map[string]interface{}{
			"type":    "kv",
			"options": map[string]interface{}{"version": "2"},
		},
		"kv1/": map[string]interface{}{
			"type":    "kv",
			"options": map[string]interface{}{},
		},
		"sys/": map[string]interface{}{
			"type":    "system",
			"options": map[string]interface{}{},
		},
	}
}

func TestListKVMounts_ReturnsOnlyKV(t *testing.T) {
	c := &Client{logical: &fakeLogicalForMounts{data: mountsData()}}
	mounts, err := c.ListKVMounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 2 {
		t.Fatalf("expected 2 KV mounts, got %d", len(mounts))
	}
}

func TestListKVMounts_DetectsVersion(t *testing.T) {
	c := &Client{logical: &fakeLogicalForMounts{data: mountsData()}}
	mounts, err := c.ListKVMounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	byPath := make(map[string]MountInfo)
	for _, m := range mounts {
		byPath[m.Path] = m
	}
	if byPath["secret"].Version != 2 {
		t.Errorf("expected secret/ to be KV v2, got v%d", byPath["secret"].Version)
	}
	if byPath["kv1"].Version != 1 {
		t.Errorf("expected kv1/ to be KV v1, got v%d", byPath["kv1"].Version)
	}
}

func TestListKVMounts_EmptyResponse(t *testing.T) {
	c := &Client{logical: &fakeLogicalForMounts{data: map[string]interface{}{}}}
	mounts, err := c.ListKVMounts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mounts) != 0 {
		t.Errorf("expected 0 mounts, got %d", len(mounts))
	}
}
