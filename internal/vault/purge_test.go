package vault

import (
	"errors"
	"testing"
)

// fakePurgeClient implements secretDeleter for testing.
type fakePurgeClient struct {
	deleted []string
	errOn   string
}

func (f *fakePurgeClient) Delete(path string) error {
	if f.errOn != "" && path == f.errOn {
		return errors.New("simulated delete error")
	}
	f.deleted = append(f.deleted, path)
	return nil
}

func (f *fakePurgeClient) List(path string) ([]string, error) {
	return nil, nil
}

func TestPurgeSecret_Success(t *testing.T) {
	client := &fakePurgeClient{}
	r := PurgeSecret(client, "secret", "app/db", false)
	if !r.Deleted {
		t.Fatal("expected Deleted=true")
	}
	if r.Error != nil {
		t.Fatalf("unexpected error: %v", r.Error)
	}
	if len(client.deleted) != 1 {
		t.Fatalf("expected 1 deletion, got %d", len(client.deleted))
	}
}

func TestPurgeSecret_DryRun(t *testing.T) {
	client := &fakePurgeClient{}
	r := PurgeSecret(client, "secret", "app/db", true)
	if !r.Skipped {
		t.Fatal("expected Skipped=true in dry-run mode")
	}
	if len(client.deleted) != 0 {
		t.Fatal("expected no actual deletions in dry-run mode")
	}
}

func TestPurgeSecret_Error(t *testing.T) {
	client := &fakePurgeClient{errOn: "secret/metadata/app/db"}
	r := PurgeSecret(client, "secret", "app/db", false)
	if r.Error == nil {
		t.Fatal("expected an error")
	}
	if r.Deleted {
		t.Fatal("Deleted should be false on error")
	}
}

func TestPurgeSecrets_Summary(t *testing.T) {
	client := &fakePurgeClient{errOn: "secret/metadata/bad/path"}
	paths := []string{"good/one", "good/two", "bad/path"}
	results, summary := PurgeSecrets(client, "secret", paths, false)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if summary.Deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", summary.Deleted)
	}
	if summary.Errors != 1 {
		t.Errorf("expected 1 error, got %d", summary.Errors)
	}
	if summary.Skipped != 0 {
		t.Errorf("expected 0 skipped, got %d", summary.Skipped)
	}
}

func TestPurgeSecrets_DryRunSummary(t *testing.T) {
	client := &fakePurgeClient{}
	paths := []string{"a", "b", "c"}
	_, summary := PurgeSecrets(client, "secret", paths, true)
	if summary.Skipped != 3 {
		t.Errorf("expected 3 skipped, got %d", summary.Skipped)
	}
	if summary.Deleted != 0 {
		t.Errorf("expected 0 deleted in dry-run, got %d", summary.Deleted)
	}
}
