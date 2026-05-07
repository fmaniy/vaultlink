package vault

import (
	"errors"
	"testing"
)

type fakeDedupeClient struct {
	data map[string]map[string]interface{}
}

func (f *fakeDedupeClient) Read(path string) (map[string]interface{}, error) {
	if d, ok := f.data[path]; ok {
		return d, nil
	}
	return nil, errors.New("not found")
}

func newDedupeClient(entries map[string]map[string]interface{}) SecretReader {
	return &fakeDedupeClient{data: entries}
}

func TestDedupeSecrets_NoDuplicates(t *testing.T) {
	c := newDedupeClient(map[string]map[string]interface{}{
		"secret/a": {"key": "alpha"},
		"secret/b": {"key": "beta"},
	})

	results := DedupeSecrets(c, []string{"secret/a", "secret/b"})
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected error for %s: %v", r.Path, r.Err)
		}
		if r.Duplicate {
			t.Errorf("expected no duplicate for %s", r.Path)
		}
	}
}

func TestDedupeSecrets_DetectsDuplicate(t *testing.T) {
	c := newDedupeClient(map[string]map[string]interface{}{
		"secret/a": {"key": "same"},
		"secret/b": {"key": "same"},
	})

	results := DedupeSecrets(c, []string{"secret/a", "secret/b"})
	if results[0].Duplicate {
		t.Error("first occurrence should not be flagged as duplicate")
	}
	if !results[1].Duplicate {
		t.Error("second occurrence should be flagged as duplicate")
	}
	if results[1].MatchPath != "secret/a" {
		t.Errorf("expected match path secret/a, got %s", results[1].MatchPath)
	}
}

func TestDedupeSecrets_ReadError(t *testing.T) {
	c := newDedupeClient(map[string]map[string]interface{}{
		"secret/a": {"key": "val"},
	})

	results := DedupeSecrets(c, []string{"secret/a", "secret/missing"})
	if results[1].Err == nil {
		t.Error("expected error for missing secret")
	}
}

func TestDedupeSecrets_MultipleGroups(t *testing.T) {
	c := newDedupeClient(map[string]map[string]interface{}{
		"secret/a": {"x": "1"},
		"secret/b": {"x": "2"},
		"secret/c": {"x": "1"},
		"secret/d": {"x": "2"},
	})

	results := DedupeSecrets(c, []string{"secret/a", "secret/b", "secret/c", "secret/d"})
	if results[2].MatchPath != "secret/a" {
		t.Errorf("expected match secret/a, got %s", results[2].MatchPath)
	}
	if results[3].MatchPath != "secret/b" {
		t.Errorf("expected match secret/b, got %s", results[3].MatchPath)
	}
}
