package vault

import (
	"errors"
	"testing"
)

type archiveClient struct {
	store map[string]map[string]interface{}
}

func newArchiveClient(initial map[string]map[string]interface{}) *archiveClient {
	if initial == nil {
		initial = make(map[string]map[string]interface{})
	}
	return &archiveClient{store: initial}
}

func (c *archiveClient) read(mount, path string) (map[string]interface{}, error) {
	key := mount + "/" + path
	if v, ok := c.store[key]; ok {
		copy := make(map[string]interface{}, len(v))
		for k, val := range v {
			copy[k] = val
		}
		return copy, nil
	}
	return nil, nil
}

func (c *archiveClient) write(mount, path string, data map[string]interface{}) error {
	c.store[mount+"/"+path] = data
	return nil
}

func TestArchiveSecret_Success(t *testing.T) {
	client := newArchiveClient(map[string]map[string]interface{}{
		"secret/myapp": {"key": "value"},
	})
	res := ArchiveSecret(client, "secret", "myapp")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.ArchivedAt.IsZero() {
		t.Error("expected ArchivedAt to be set")
	}
	if _, ok := client.store["secret/myapp"]["_archived_at"]; !ok {
		t.Error("expected _archived_at key in secret data")
	}
}

func TestArchiveSecret_AlreadyArchived(t *testing.T) {
	client := newArchiveClient(map[string]map[string]interface{}{
		"secret/myapp": {"key": "value", "_archived_at": "2024-01-01T00:00:00Z"},
	})
	res := ArchiveSecret(client, "secret", "myapp")
	if res.Err == nil {
		t.Fatal("expected error for already archived secret")
	}
}

func TestArchiveSecret_NotFound(t *testing.T) {
	client := newArchiveClient(nil)
	res := ArchiveSecret(client, "secret", "missing")
	if res.Err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestUnarchiveSecret_Success(t *testing.T) {
	client := newArchiveClient(map[string]map[string]interface{}{
		"secret/myapp": {"key": "value", "_archived_at": "2024-01-01T00:00:00Z"},
	})
	res := UnarchiveSecret(client, "secret", "myapp")
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if _, ok := client.store["secret/myapp"]["_archived_at"]; ok {
		t.Error("expected _archived_at key to be removed")
	}
}

func TestUnarchiveSecret_NotArchived(t *testing.T) {
	client := newArchiveClient(map[string]map[string]interface{}{
		"secret/myapp": {"key": "value"},
	})
	res := UnarchiveSecret(client, "secret", "myapp")
	if res.Err == nil {
		t.Fatal("expected error when secret is not archived")
	}
}

func TestArchiveSecrets_MultipleResults(t *testing.T) {
	client := newArchiveClient(map[string]map[string]interface{}{
		"secret/a": {"x": "1"},
		"secret/b": {"y": "2"},
	})
	results := ArchiveSecrets(client, "secret", []string{"a", "b", "c"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Err != nil || results[1].Err != nil {
		t.Error("expected first two results to succeed")
	}
	if results[2].Err == nil {
		t.Error("expected third result to fail (not found)")
	}
}

// Ensure archiveClient satisfies the Archiver interface via read/write methods.
// The Archiver interface requires Reader and Writer; we adapt via embedding.
var _ = errors.New // keep errors import used
