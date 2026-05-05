package vault

import (
	"errors"
	"testing"
)

func newAnnotateClient(store map[string]map[string]interface{}) *Client {
	return clientWithFakeLogical(store)
}

func TestAnnotateSecret_AddsNewAnnotation(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/app": {"password": "s3cr3t"},
	}
	c := newAnnotateClient(store)

	res := AnnotateSecret(c, "secret/data/app", "owner", "team-a")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if !res.Updated {
		t.Fatal("expected Updated=true")
	}

	data, _ := ReadSecret(c, "secret/data/app")
	if data["_annotations"] != "owner=team-a" {
		t.Errorf("unexpected annotations: %v", data["_annotations"])
	}
}

func TestAnnotateSecret_MergesWithExisting(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/app": {"password": "s3cr3t", "_annotations": "env=prod"},
	}
	c := newAnnotateClient(store)

	res := AnnotateSecret(c, "secret/data/app", "owner", "team-b")

	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}

	data, _ := ReadSecret(c, "secret/data/app")
	annotStr, _ := data["_annotations"].(string)
	if !contains(annotStr, "env=prod") || !contains(annotStr, "owner=team-b") {
		t.Errorf("annotations not merged correctly: %v", annotStr)
	}
}

func TestAnnotateSecret_OverwritesExistingKey(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/app": {"_annotations": "env=staging"},
	}
	c := newAnnotateClient(store)

	AnnotateSecret(c, "secret/data/app", "env", "prod")

	data, _ := ReadSecret(c, "secret/data/app")
	if data["_annotations"] != "env=prod" {
		t.Errorf("expected env=prod, got %v", data["_annotations"])
	}
}

func TestAnnotateSecret_ReadError(t *testing.T) {
	c := newAnnotateClient(map[string]map[string]interface{}{})

	res := AnnotateSecret(c, "secret/data/missing", "key", "val")

	if res.Err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestAnnotateSecrets_MultipleResults(t *testing.T) {
	store := map[string]map[string]interface{}{
		"secret/data/a": {"x": "1"},
		"secret/data/b": {"y": "2"},
	}
	c := newAnnotateClient(store)

	results := AnnotateSecrets(c, []string{"secret/data/a", "secret/data/b"}, "team", "ops")

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Err)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

var _ = errors.New // keep import
