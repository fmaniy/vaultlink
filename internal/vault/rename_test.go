package vault

import (
	"errors"
	"testing"
)

// renameClientSetup returns a fakeLogical pre-seeded with one secret.
func renameClientSetup() (*Client, *fakeLogical) {
	fl := &fakeLogical{
		readData: map[string]map[string]interface{}{
			"secret/data/old": {"key": "value"},
		},
		writtenData: map[string]map[string]interface{}{},
		deletedPaths: []string{},
	}
	c := clientWithFakeLogical(fl)
	return c, fl
}

func TestRenameSecret_Success(t *testing.T) {
	c, fl := renameClientSetup()

	if err := RenameSecret(c, "secret", "old", "new"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := fl.writtenData["secret/data/new"]; !ok {
		t.Error("expected secret to be written to destination path")
	}

	if len(fl.deletedPaths) == 0 || fl.deletedPaths[0] != "secret/metadata/old" {
		t.Errorf("expected source path to be deleted, got %v", fl.deletedPaths)
	}
}

func TestRenameSecret_SamePath(t *testing.T) {
	c, _ := renameClientSetup()

	err := RenameSecret(c, "secret", "old", "old")
	if err == nil {
		t.Fatal("expected error for identical src and dst, got nil")
	}
}

func TestRenameSecret_ReadError(t *testing.T) {
	fl := &fakeLogical{
		readErr:      errors.New("read failed"),
		writtenData:  map[string]map[string]interface{}{},
		deletedPaths: []string{},
	}
	c := clientWithFakeLogical(fl)

	err := RenameSecret(c, "secret", "missing", "new")
	if err == nil {
		t.Fatal("expected read error, got nil")
	}
}

func TestRenameSecret_WriteError(t *testing.T) {
	fl := &fakeLogical{
		readData: map[string]map[string]interface{}{
			"secret/data/old": {"k": "v"},
		},
		writeErr:     errors.New("write failed"),
		writtenData:  map[string]map[string]interface{}{},
		deletedPaths: []string{},
	}
	c := clientWithFakeLogical(fl)

	err := RenameSecret(c, "secret", "old", "new")
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestRenameSecrets_MultipleEntries(t *testing.T) {
	fl := &fakeLogical{
		readData: map[string]map[string]interface{}{
			"secret/data/a": {"x": "1"},
			"secret/data/b": {"y": "2"},
		},
		writtenData:  map[string]map[string]interface{}{},
		deletedPaths: []string{},
	}
	c := clientWithFakeLogical(fl)

	paths := map[string]string{"a": "a_new", "b": "b_new"}
	if err := RenameSecrets(c, "secret", paths); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fl.writtenData) != 2 {
		t.Errorf("expected 2 written secrets, got %d", len(fl.writtenData))
	}
}
