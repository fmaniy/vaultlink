package vault

import (
	"errors"
	"testing"
)

func newPatchClient(data map[string]map[string]interface{}) *Client {
	return clientWithFakeLogical(&fakeLogical{data: data})
}

func TestPatchSecret_UpdatesExistingKey(t *testing.T) {
	c := newPatchClient(map[string]map[string]interface{}{
		"secret/data/app": {"db_pass": "old"},
	})
	res := PatchSecret(c, "secret/data/app", "db_pass", "new", false)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.OldVal != "old" {
		t.Errorf("expected OldVal=old, got %q", res.OldVal)
	}
	if res.NewVal != "new" {
		t.Errorf("expected NewVal=new, got %q", res.NewVal)
	}
	if res.Skipped {
		t.Error("expected not skipped")
	}
}

func TestPatchSecret_SkipsWhenKeyMissingAndCreateFalse(t *testing.T) {
	c := newPatchClient(map[string]map[string]interface{}{
		"secret/data/app": {"other": "val"},
	})
	res := PatchSecret(c, "secret/data/app", "db_pass", "new", false)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Skipped {
		t.Error("expected skipped=true")
	}
}

func TestPatchSecret_CreatesKeyWhenAllowed(t *testing.T) {
	c := newPatchClient(map[string]map[string]interface{}{
		"secret/data/app": {"other": "val"},
	})
	res := PatchSecret(c, "secret/data/app", "db_pass", "created", true)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Skipped {
		t.Error("expected not skipped")
	}
}

func TestPatchSecret_SecretNotFound(t *testing.T) {
	c := newPatchClient(map[string]map[string]interface{}{})
	res := PatchSecret(c, "secret/data/missing", "key", "val", false)
	if res.Err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestPatchSecret_ReadError(t *testing.T) {
	fl := &fakeLogical{err: errors.New("vault down")}
	c := clientWithFakeLogical(fl)
	res := PatchSecret(c, "secret/data/app", "key", "val", false)
	if res.Err == nil {
		t.Fatal("expected read error")
	}
}

func TestPatchSecrets_ReturnsAllResults(t *testing.T) {
	c := newPatchClient(map[string]map[string]interface{}{
		"secret/data/a": {"k": "v1"},
		"secret/data/b": {"k": "v2"},
	})
	results := PatchSecrets(c, []string{"secret/data/a", "secret/data/b"}, "k", "new", false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Err)
		}
	}
}
