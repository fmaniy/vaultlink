package vault

import (
	"testing"
)

func TestDiffSecrets_AllMatch(t *testing.T) {
	src := map[string]interface{}{"foo": "bar", "baz": "qux"}
	dst := map[string]interface{}{"foo": "bar", "baz": "qux"}

	results := DiffSecrets("secret/app", src, dst)
	for _, r := range results {
		if r.Status != DiffStatusMatch {
			t.Errorf("expected match for key %s, got %s", r.Key, r.Status)
		}
	}
}

func TestDiffSecrets_MissingInDst(t *testing.T) {
	src := map[string]interface{}{"foo": "bar"}
	dst := map[string]interface{}{}

	results := DiffSecrets("secret/app", src, dst)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != DiffStatusMissing {
		t.Errorf("expected missing, got %s", results[0].Status)
	}
}

func TestDiffSecrets_ExtraInDst(t *testing.T) {
	src := map[string]interface{}{}
	dst := map[string]interface{}{"extra": "val"}

	results := DiffSecrets("secret/app", src, dst)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != DiffStatusExtra {
		t.Errorf("expected extra, got %s", results[0].Status)
	}
}

func TestDiffSecrets_Mismatch(t *testing.T) {
	src := map[string]interface{}{"key": "val1"}
	dst := map[string]interface{}{"key": "val2"}

	results := DiffSecrets("secret/app", src, dst)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != DiffStatusMismatch {
		t.Errorf("expected mismatch, got %s", results[0].Status)
	}
	if results[0].SrcVal != "val1" || results[0].DstVal != "val2" {
		t.Errorf("unexpected values: src=%s dst=%s", results[0].SrcVal, results[0].DstVal)
	}
}

func TestDiffSecrets_PathIsPreserved(t *testing.T) {
	src := map[string]interface{}{"k": "v"}
	dst := map[string]interface{}{"k": "v"}

	results := DiffSecrets("secret/myapp/config", src, dst)
	if len(results) == 0 {
		t.Fatal("expected results")
	}
	if results[0].Path != "secret/myapp/config" {
		t.Errorf("unexpected path: %s", results[0].Path)
	}
}
