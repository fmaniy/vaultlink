package vault

import (
	"errors"
	"testing"
)

// fakeTouchClient implements touchWriter for testing.
type fakeTouchClient struct {
	store   map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (f *fakeTouchClient) Read(path string) (map[string]interface{}, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	v, ok := f.store[path]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (f *fakeTouchClient) Write(path string, data map[string]interface{}) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	f.store[path] = data
	return nil
}

func newTouchClient(mount, path string, data map[string]interface{}) *fakeTouchClient {
	fullPath := mount + "/data/" + path
	return &fakeTouchClient{
		store: map[string]map[string]interface{}{
			fullPath: {"data": data},
		},
	}
}

func TestTouchSecret_Success(t *testing.T) {
	client := newTouchClient("secret", "myapp/db", map[string]interface{}{"password": "s3cr3t"})
	res := TouchSecret(client, "secret", "myapp/db")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if !res.Updated {
		t.Error("expected Updated=true")
	}
	// Verify _touched_at was written
	stored := client.store["secret/data/myapp/db"]
	inner, _ := stored["data"].(map[string]interface{})
	if inner["_touched_at"] == nil {
		t.Error("expected _touched_at to be set")
	}
}

func TestTouchSecret_NotFound(t *testing.T) {
	client := &fakeTouchClient{store: map[string]map[string]interface{}{}}
	res := TouchSecret(client, "secret", "missing/path")
	if res.Error == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestTouchSecret_ReadError(t *testing.T) {
	client := &fakeTouchClient{readErr: errors.New("vault unavailable")}
	res := TouchSecret(client, "secret", "myapp/db")
	if res.Error == nil {
		t.Fatal("expected read error")
	}
}

func TestTouchSecret_WriteError(t *testing.T) {
	client := newTouchClient("secret", "myapp/db", map[string]interface{}{"key": "val"})
	client.writeErr = errors.New("permission denied")
	res := TouchSecret(client, "secret", "myapp/db")
	if res.Error == nil {
		t.Fatal("expected write error")
	}
}

func TestTouchSecrets_MultipleResults(t *testing.T) {
	client := newTouchClient("secret", "a", map[string]interface{}{"x": "1"})
	client.store["secret/data/b"] = map[string]interface{}{"data": map[string]interface{}{"y": "2"}}

	results := TouchSecrets(client, "secret", []string{"a", "b", "missing"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if !results[0].Updated || !results[1].Updated {
		t.Error("expected first two results to be updated")
	}
	if results[2].Error == nil {
		t.Error("expected error for missing secret")
	}
}
