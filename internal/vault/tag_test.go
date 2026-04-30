package vault

import (
	"errors"
	"testing"
)

func newTagClient(data map[string]map[string]interface{}) *Client {
	return clientWithFakeLogical(&fakeLogical{secrets: data})
}

func TestTagSecret_AddsNewTags(t *testing.T) {
	client := newTagClient(map[string]map[string]interface{}{
		"secret/data/app": {"DB_PASS": "hunter2"},
	})

	res := TagSecret(client, "secret", "app", []string{"prod", "critical"})
	if !res.Success {
		t.Fatalf("expected success, got err: %v", res.Err)
	}
	if len(res.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(res.Tags))
	}
}

func TestTagSecret_MergesWithExisting(t *testing.T) {
	client := newTagClient(map[string]map[string]interface{}{
		"secret/data/app": {"key": "val", tagMetaKey: "prod"},
	})

	res := TagSecret(client, "secret", "app", []string{"critical"})
	if !res.Success {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Tags) != 2 {
		t.Errorf("expected 2 tags after merge, got %d: %v", len(res.Tags), res.Tags)
	}
}

func TestTagSecret_NoDuplicates(t *testing.T) {
	client := newTagClient(map[string]map[string]interface{}{
		"secret/data/app": {"key": "val", tagMetaKey: "prod,critical"},
	})

	res := TagSecret(client, "secret", "app", []string{"prod"})
	if len(res.Tags) != 2 {
		t.Errorf("expected no duplicate, got %d tags: %v", len(res.Tags), res.Tags)
	}
}

func TestTagSecret_ReadError(t *testing.T) {
	client := clientWithFakeLogical(&fakeLogical{
		readErr: errors.New("permission denied"),
	})
	res := TagSecret(client, "secret", "app", []string{"prod"})
	if res.Success {
		t.Fatal("expected failure")
	}
	if res.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestUntagSecret_RemovesTag(t *testing.T) {
	client := newTagClient(map[string]map[string]interface{}{
		"secret/data/app": {"key": "val", tagMetaKey: "prod,critical,debug"},
	})

	res := UntagSecret(client, "secret", "app", []string{"debug"})
	if !res.Success {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if len(res.Tags) != 2 {
		t.Errorf("expected 2 tags remaining, got %d: %v", len(res.Tags), res.Tags)
	}
	for _, tag := range res.Tags {
		if tag == "debug" {
			t.Error("debug tag should have been removed")
		}
	}
}

func TestTagSecrets_MultiPath(t *testing.T) {
	client := newTagClient(map[string]map[string]interface{}{
		"secret/data/a": {"k": "v"},
		"secret/data/b": {"k": "v"},
	})

	results := TagSecrets(client, "secret", []string{"a", "b"}, []string{"env:prod"})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("path %s failed: %v", r.Path, r.Err)
		}
	}
}
