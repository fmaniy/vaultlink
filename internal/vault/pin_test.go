package vault

import (
	"errors"
	"testing"
)

type pinFakeLogical struct {
	store map[string]map[string]interface{}
	readErr  error
	writeErr error
}

func (f *pinFakeLogical) Read(path string) (*fakeSecret, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	data, ok := f.store[path]
	if !ok {
		return nil, nil
	}
	return &fakeSecret{Data: map[string]interface{}{"data": toIfaceMap(data)}}, nil
}

func (f *pinFakeLogical) Write(path string, data map[string]interface{}) (*fakeSecret, error) {
	if f.writeErr != nil {
		return nil, f.writeErr
	}
	if inner, ok := data["data"].(map[string]interface{}); ok {
		f.store[path] = toStringMap(inner)
	}
	return &fakeSecret{}, nil
}

func newPinClient(store map[string]map[string]interface{}) *Client {
	return clientWithFakeLogical(&pinFakeLogical{store: store})
}

func TestPinSecret_Success(t *testing.T) {
	client := newPinClient(map[string]map[string]interface{}{
		"secret/data/app/db": {"password": "s3cr3t"},
	})
	res := PinSecret(client, "secret", "app/db", "42")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if !res.Pinned {
		t.Error("expected Pinned to be true")
	}
	if res.Version != "42" {
		t.Errorf("expected version 42, got %s", res.Version)
	}
}

func TestPinSecret_NotFound(t *testing.T) {
	client := newPinClient(map[string]map[string]interface{}{})
	res := PinSecret(client, "secret", "app/missing", "1")
	if res.Error == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestPinSecret_ReadError(t *testing.T) {
	fl := &pinFakeLogical{store: map[string]map[string]interface{}{}, readErr: errors.New("vault down")}
	client := clientWithFakeLogical(fl)
	res := PinSecret(client, "secret", "app/db", "1")
	if res.Error == nil {
		t.Fatal("expected read error")
	}
}

func TestUnpinSecret_RemovesMetadata(t *testing.T) {
	client := newPinClient(map[string]map[string]interface{}{
		"secret/data/app/db": {
			"password":       "s3cr3t",
			pinVersionKey:    "5",
			pinTimestampKey:  "2024-01-01T00:00:00Z",
		},
	})
	res := UnpinSecret(client, "secret", "app/db")
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if res.Pinned {
		t.Error("expected Pinned to be false after unpin")
	}
}

func TestPinSecrets_MultipleResults(t *testing.T) {
	client := newPinClient(map[string]map[string]interface{}{
		"secret/data/a": {"k": "v"},
		"secret/data/b": {"k": "v"},
	})
	results := PinSecrets(client, "secret", []string{"a", "b"}, "3")
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error for %s: %v", r.Path, r.Error)
		}
	}
}
