package vault

import (
	"errors"
	"testing"
)

func cloneSetup() (*fakeLogical, *fakeLogical) {
	src := &fakeLogical{
		readData: map[string]map[string]interface{}{
			"secret/data/app/db": {"password": "s3cr3t"},
		},
	}
	dst := &fakeLogical{
		readData: map[string]map[string]interface{}{},
	}
	return src, dst
}

func TestCloneSecret_Success(t *testing.T) {
	src, dst := cloneSetup()
	res := CloneSecret(src, dst, "secret", "secret", "app/db", "app/db", false)
	if res.Status != "cloned" {
		t.Fatalf("expected cloned, got %s", res.Status)
	}
}

func TestCloneSecret_SkippedWhenExists(t *testing.T) {
	src, dst := cloneSetup()
	dst.readData["secret/data/app/db"] = map[string]interface{}{"password": "old"}
	res := CloneSecret(src, dst, "secret", "secret", "app/db", "app/db", false)
	if res.Status != "skipped" {
		t.Fatalf("expected skipped, got %s", res.Status)
	}
}

func TestCloneSecret_OverwriteReplaces(t *testing.T) {
	src, dst := cloneSetup()
	dst.readData["secret/data/app/db"] = map[string]interface{}{"password": "old"}
	res := CloneSecret(src, dst, "secret", "secret", "app/db", "app/db", true)
	if res.Status != "cloned" {
		t.Fatalf("expected cloned, got %s", res.Status)
	}
}

func TestCloneSecret_ReadError(t *testing.T) {
	src := &fakeLogical{readErr: errors.New("vault down")}
	dst := &fakeLogical{}
	res := CloneSecret(src, dst, "secret", "secret", "app/db", "app/db", false)
	if res.Status != "error" {
		t.Fatalf("expected error, got %s", res.Status)
	}
}

func TestCloneSecret_WriteError(t *testing.T) {
	src, dst := cloneSetup()
	dst.writeErr = errors.New("permission denied")
	res := CloneSecret(src, dst, "secret", "secret", "app/db", "app/db", false)
	if res.Status != "error" {
		t.Fatalf("expected error, got %s", res.Status)
	}
}

func TestCloneSecrets_MultipleResults(t *testing.T) {
	src := &fakeLogical{
		readData: map[string]map[string]interface{}{
			"secret/data/a": {"k": "v"},
			"secret/data/b": {"k": "v"},
		},
	}
	dst := &fakeLogical{readData: map[string]map[string]interface{}{}}
	results := CloneSecrets(src, dst, "secret", "secret", []string{"a", "b"}, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "cloned" {
			t.Errorf("expected cloned, got %s for %s", r.Status, r.SourcePath)
		}
	}
}
