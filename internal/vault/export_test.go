package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

type exportFakeLogical struct {
	data map[string]map[string]interface{}
}

func (f *exportFakeLogical) Read(path string) (*fakeSecret, error) {
	data, ok := f.data[path]
	if !ok {
		return nil, errors.New("not found")
	}
	return &fakeSecret{Data: map[string]interface{}{"data": toIface(data)}}, nil
}

func (f *exportFakeLogical) Write(path string, data map[string]interface{}) (*fakeSecret, error) {
	return nil, nil
}

func toIface(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func TestExportSecrets_Success(t *testing.T) {
	client := clientWithFakeLogical(&exportFakeLogical{
		data: map[string]map[string]interface{}{
			"secret/data/app/db": {"password": "s3cr3t"},
			"secret/data/app/api": {"key": "abc123"},
		},
	})

	var buf bytes.Buffer
	results, err := ExportSecrets(client, "secret", []string{"app/db", "app/api"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Success {
			t.Errorf("expected success for path %q, got error: %s", r.Path, r.Error)
		}
	}

	var out map[string]map[string]string
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if out["app/db"]["password"] != "s3cr3t" {
		t.Errorf("expected password=s3cr3t, got %q", out["app/db"]["password"])
	}
}

func TestExportSecrets_PartialFailure(t *testing.T) {
	client := clientWithFakeLogical(&exportFakeLogical{
		data: map[string]map[string]interface{}{
			"secret/data/app/db": {"password": "s3cr3t"},
		},
	})

	var buf bytes.Buffer
	results, err := ExportSecrets(client, "secret", []string{"app/db", "app/missing"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Success {
		t.Errorf("expected first result to succeed")
	}
	if results[1].Success {
		t.Errorf("expected second result to fail")
	}
	if results[1].Error == "" {
		t.Errorf("expected non-empty error message for failed result")
	}
}
