package vault

import (
	"errors"
	"testing"
)

type promoteClient struct {
	data   map[string]map[string]interface{}
	writen map[string]map[string]interface{}
	readErr  error
	writeErr error
}

func (c *promoteClient) ReadSecret(path string) (map[string]interface{}, error) {
	if c.readErr != nil {
		return nil, c.readErr
	}
	v, ok := c.data[path]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (c *promoteClient) WriteSecret(path string, data map[string]interface{}) error {
	if c.writeErr != nil {
		return c.writeErr
	}
	if c.writen == nil {
		c.writen = make(map[string]map[string]interface{})
	}
	c.writen[path] = data
	return nil
}

func TestPromoteSecret_Success(t *testing.T) {
	src := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "s3cr3t"}}}
	dst := &promoteClient{data: map[string]map[string]interface{}{}}

	r := PromoteSecret(src, dst, "app/db", false)
	if !r.Success {
		t.Fatalf("expected success, got error: %v", r.Error)
	}
	if dst.writen["app/db"]["pass"] != "s3cr3t" {
		t.Errorf("expected secret to be written to dst")
	}
}

func TestPromoteSecret_NoOverwrite(t *testing.T) {
	src := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "new"}}}
	dst := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "old"}}}

	r := PromoteSecret(src, dst, "app/db", false)
	if r.Success {
		t.Fatal("expected failure when secret exists and overwrite=false")
	}
}

func TestPromoteSecret_OverwriteReplaces(t *testing.T) {
	src := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "new"}}}
	dst := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "old"}}}

	r := PromoteSecret(src, dst, "app/db", true)
	if !r.Success {
		t.Fatalf("expected success with overwrite, got: %v", r.Error)
	}
}

func TestPromoteSecret_ReadError(t *testing.T) {
	src := &promoteClient{readErr: errors.New("vault down")}
	dst := &promoteClient{data: map[string]map[string]interface{}{}}

	r := PromoteSecret(src, dst, "app/db", false)
	if r.Success {
		t.Fatal("expected failure on read error")
	}
}

func TestPromoteSecret_WriteError(t *testing.T) {
	src := &promoteClient{data: map[string]map[string]interface{}{"app/db": {"pass": "s3cr3t"}}}
	dst := &promoteClient{writeErr: errors.New("permission denied")}

	r := PromoteSecret(src, dst, "app/db", false)
	if r.Success {
		t.Fatal("expected failure on write error")
	}
	if r.Error == nil {
		t.Fatal("expected non-nil error on write failure")
	}
}

func TestPromoteSecrets_MultipleResults(t *testing.T) {
	src := &promoteClient{data: map[string]map[string]interface{}{
		"a": {"k": "v"},
		"b": {"k": "v"},
	}}
	dst := &promoteClient{data: map[string]map[string]interface{}{}}

	results := PromoteSecrets(src, dst, []string{"a", "b"}, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for path %s: %v", r.Path, r.Error)
		}
	}
}
