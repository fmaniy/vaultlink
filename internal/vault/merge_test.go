package vault

import (
	"errors"
	"testing"
)

type fakeMergeClient struct {
	store    map[string]map[string]interface{}
	writeErr error
}

func (f *fakeMergeClient) ReadSecret(mount, path string) (map[string]interface{}, error) {
	key := mount + "/" + path
	if d, ok := f.store[key]; ok {
		return d, nil
	}
	return nil, errors.New("not found")
}

func (f *fakeMergeClient) WriteSecret(mount, path string, data map[string]interface{}) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.store[mount+"/"+path] = data
	return nil
}

func newMergeClient() *fakeMergeClient {
	return &fakeMergeClient{store: map[string]map[string]interface{}{}}
}

func TestMergeSecret_Success(t *testing.T) {
	c := newMergeClient()
	c.store["src/app"] = map[string]interface{}{"a": "1", "b": "2"}
	c.store["dst/app"] = map[string]interface{}{"b": "old", "c": "3"}

	res := MergeSecret(c, "src", "app", "dst", "app", MergeOptions{Overwrite: true})

	if res.Status != "merged" {
		t.Fatalf("expected merged, got %s", res.Status)
	}
	if c.store["dst/app"]["a"] != "1" {
		t.Error("expected key a from src")
	}
	if c.store["dst/app"]["c"] != "3" {
		t.Error("expected key c preserved from dst")
	}
	if len(res.Conflicts) != 1 || res.Conflicts[0] != "b" {
		t.Errorf("expected conflict on b, got %v", res.Conflicts)
	}
}

func TestMergeSecret_NoOverwrite(t *testing.T) {
	c := newMergeClient()
	c.store["src/app"] = map[string]interface{}{"key": "new"}
	c.store["dst/app"] = map[string]interface{}{"key": "old"}

	res := MergeSecret(c, "src", "app", "dst", "app", MergeOptions{Overwrite: false})

	if res.Status != "merged" {
		t.Fatalf("expected merged, got %s", res.Status)
	}
	if c.store["dst/app"]["key"] != "old" {
		t.Error("expected original value preserved when overwrite=false")
	}
}

func TestMergeSecret_DryRun(t *testing.T) {
	c := newMergeClient()
	c.store["src/app"] = map[string]interface{}{"x": "1"}

	res := MergeSecret(c, "src", "app", "dst", "app", MergeOptions{DryRun: true})

	if res.Status != "skipped" {
		t.Fatalf("expected skipped, got %s", res.Status)
	}
	if _, exists := c.store["dst/app"]; exists {
		t.Error("dry run should not write")
	}
}

func TestMergeSecret_ReadError(t *testing.T) {
	c := newMergeClient() // src key absent

	res := MergeSecret(c, "src", "missing", "dst", "missing", MergeOptions{})

	if res.Status != "error" {
		t.Fatalf("expected error, got %s", res.Status)
	}
}

func TestMergeSecret_WriteError(t *testing.T) {
	c := newMergeClient()
	c.store["src/app"] = map[string]interface{}{"k": "v"}
	c.writeErr = errors.New("vault unavailable")

	res := MergeSecret(c, "src", "app", "dst", "app", MergeOptions{})

	if res.Status != "error" {
		t.Fatalf("expected error, got %s", res.Status)
	}
}

func TestMergeSecrets_MultipleResults(t *testing.T) {
	c := newMergeClient()
	c.store["src/a"] = map[string]interface{}{"k": "1"}
	c.store["src/b"] = map[string]interface{}{"k": "2"}

	results := MergeSecrets(c, "src", []string{"a", "b"}, "dst", MergeOptions{})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "merged" {
			t.Errorf("expected merged, got %s for %s", r.Status, r.Path)
		}
	}
}
