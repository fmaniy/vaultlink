package vault

import (
	"errors"
	"testing"
)

type fakeSnapClient struct {
	listResult []string
	listErr    error
	readData   map[string]map[string]interface{}
	readErr    map[string]error
}

func (f *fakeSnapClient) List(path string) ([]string, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listResult, nil
}

func (f *fakeSnapClient) Read(path string) (map[string]interface{}, error) {
	if err, ok := f.readErr[path]; ok {
		return nil, err
	}
	return f.readData[path], nil
}

func TestSnapshotSecrets_Success(t *testing.T) {
	client := &fakeSnapClient{
		listResult: []string{"db/password", "app/token"},
		readData: map[string]map[string]interface{}{
			"db/password": {"value": "secret1"},
			"app/token":   {"value": "secret2"},
		},
	}
	entries, results, err := SnapshotSecrets(client, "secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	for _, r := range results {
		if !r.OK {
			t.Errorf("expected OK for %s", r.Path)
		}
	}
}

func TestSnapshotSecrets_ListError(t *testing.T) {
	client := &fakeSnapClient{listErr: errors.New("permission denied")}
	_, _, err := SnapshotSecrets(client, "secret", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSnapshotSecrets_PartialReadError(t *testing.T) {
	client := &fakeSnapClient{
		listResult: []string{"good", "bad"},
		readData:   map[string]map[string]interface{}{"good": {"k": "v"}},
		readErr:    map[string]error{"bad": errors.New("not found")},
	}
	entries, results, err := SnapshotSecrets(client, "secret", "")
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	var failCount int
	for _, r := range results {
		if !r.OK {
			failCount++
		}
	}
	if failCount != 1 {
		t.Errorf("expected 1 failure, got %d", failCount)
	}
}
