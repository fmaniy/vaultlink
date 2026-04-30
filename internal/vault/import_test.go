package vault

import (
	"encoding/json"
	"os"
	"testing"
)

func writeImportFile(t *testing.T, secrets map[string]map[string]interface{}) string {
	t.Helper()
	data, err := json.Marshal(secrets)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	f, err := os.CreateTemp(t.TempDir(), "import-*.json")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestImportSecrets_Success(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"app/db": {"password": "s3cr3t"},
	}
	file := writeImportFile(t, secrets)

	store := map[string]map[string]interface{}{}
	client := clientWithFakeLogical(store)

	results, err := ImportSecrets(client, "secret", file, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "imported" {
		t.Errorf("expected status 'imported', got %q", results[0].Status)
	}
}

func TestImportSecrets_SkipExisting(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"app/db": {"password": "s3cr3t"},
	}
	file := writeImportFile(t, secrets)

	store := map[string]map[string]interface{}{
		"secret/data/app/db": {"data": map[string]interface{}{"password": "old"}},
	}
	client := clientWithFakeLogical(store)

	results, err := ImportSecrets(client, "secret", file, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != "skipped" {
		t.Errorf("expected status 'skipped', got %q", results[0].Status)
	}
}

func TestImportSecrets_OverwriteReplaces(t *testing.T) {
	secrets := map[string]map[string]interface{}{
		"app/db": {"password": "new"},
	}
	file := writeImportFile(t, secrets)

	store := map[string]map[string]interface{}{
		"secret/data/app/db": {"data": map[string]interface{}{"password": "old"}},
	}
	client := clientWithFakeLogical(store)

	results, err := ImportSecrets(client, "secret", file, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].Status != "imported" {
		t.Errorf("expected status 'imported', got %q", results[0].Status)
	}
}

func TestImportSecrets_MissingFile(t *testing.T) {
	store := map[string]map[string]interface{}{}
	client := clientWithFakeLogical(store)

	_, err := ImportSecrets(client, "secret", "/nonexistent/path.json", false)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
