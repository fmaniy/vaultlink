package vault

import (
	"errors"
	"testing"
)

type fakeTrimClient struct {
	data map[string]map[string]interface{}
	written map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (f *fakeTrimClient) Logical() logicalClient {
	return &fakeLogical{
		readFn: func(path string) (map[string]interface{}, error) {
			if f.readErr != nil {
				return nil, f.readErr
			}
			v, ok := f.data[path]
			if !ok {
				return nil, nil
			}
			out := make(map[string]interface{}, len(v))
			for k, val := range v {
				out[k] = val
			}
			return out, nil
		},
		writeFn: func(path string, data map[string]interface{}) error {
			if f.writeErr != nil {
				return f.writeErr
			}
			if f.written == nil {
				f.written = map[string]map[string]interface{}{}
			}
			f.written[path] = data
			return nil
		},
	}
}

func TestTrimSecret_TrimsWhitespace(t *testing.T) {
	client := &fakeTrimClient{
		data: map[string]map[string]interface{}{
			"secret/data/app/key": {"token": "  abc  ", "name": "no-trim"},
		},
	}
	res := TrimSecret(client, "secret", "app/key", false)
	if res.Err != nil {
		t.Fatalf("unexpected error: %v", res.Err)
	}
	if res.Trimmed != 1 {
		t.Errorf("expected 1 trimmed field, got %d", res.Trimmed)
	}
	if client.written["secret/data/app/key"]["token"] != "abc" {
		t.Errorf("expected trimmed value 'abc'")
	}
}

func TestTrimSecret_DryRunDoesNotWrite(t *testing.T) {
	client := &fakeTrimClient{
		data: map[string]map[string]interface{}{
			"secret/data/app/key": {"val": "  hello "},
		},
	}
	res := TrimSecret(client, "secret", "app/key", true)
	if res.Trimmed != 1 {
		t.Errorf("expected 1 trimmed field, got %d", res.Trimmed)
	}
	if client.written != nil {
		t.Error("dry-run should not write")
	}
}

func TestTrimSecret_NothingToTrim(t *testing.T) {
	client := &fakeTrimClient{
		data: map[string]map[string]interface{}{
			"secret/data/app/key": {"val": "clean"},
		},
	}
	res := TrimSecret(client, "secret", "app/key", false)
	if res.Trimmed != 0 {
		t.Errorf("expected 0 trimmed fields, got %d", res.Trimmed)
	}
	if client.written != nil {
		t.Error("should not write when nothing changed")
	}
}

func TestTrimSecret_ReadError(t *testing.T) {
	client := &fakeTrimClient{readErr: errors.New("vault down")}
	res := TrimSecret(client, "secret", "app/key", false)
	if res.Err == nil {
		t.Fatal("expected error")
	}
}

func TestTrimSecrets_MultipleResults(t *testing.T) {
	client := &fakeTrimClient{
		data: map[string]map[string]interface{}{
			"secret/data/a": {"x": " v "},
			"secret/data/b": {"y": "clean"},
		},
	}
	results := TrimSecrets(client, "secret", []string{"a", "b"}, false)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Trimmed != 1 {
		t.Errorf("first path: expected 1 trimmed")
	}
	if results[1].Trimmed != 0 {
		t.Errorf("second path: expected 0 trimmed")
	}
}
