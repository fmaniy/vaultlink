package vault

import (
	"errors"
	"testing"
)

type fakeValidateClient struct {
	data map[string]map[string]interface{}
}

func (f *fakeValidateClient) ReadSecret(mount, path string) (map[string]interface{}, error) {
	key := mount + "/" + path
	if d, ok := f.data[key]; ok {
		return d, nil
	}
	return nil, errors.New("secret not found")
}

func TestValidateSecret_AllKeysPresent(t *testing.T) {
	c := &fakeValidateClient{
		data: map[string]map[string]interface{}{
			"secret/myapp/db": {"host": "localhost", "port": "5432", "pass": "s3cr3t"},
		},
	}
	res := ValidateSecret(c, "secret", "myapp/db", []string{"host", "port", "pass"})
	if res.Status != "ok" {
		t.Fatalf("expected ok, got %s: %v", res.Status, res.Err)
	}
	if len(res.Missing) != 0 {
		t.Fatalf("expected no missing keys, got %v", res.Missing)
	}
}

func TestValidateSecret_MissingKeys(t *testing.T) {
	c := &fakeValidateClient{
		data: map[string]map[string]interface{}{
			"secret/myapp/db": {"host": "localhost"},
		},
	}
	res := ValidateSecret(c, "secret", "myapp/db", []string{"host", "port", "pass"})
	if res.Status != "missing_keys" {
		t.Fatalf("expected missing_keys, got %s", res.Status)
	}
	if len(res.Missing) != 2 {
		t.Fatalf("expected 2 missing keys, got %v", res.Missing)
	}
}

func TestValidateSecret_NotFound(t *testing.T) {
	c := &fakeValidateClient{data: map[string]map[string]interface{}{}}
	res := ValidateSecret(c, "secret", "missing/path", []string{"key"})
	if res.Status != "not_found" {
		t.Fatalf("expected not_found, got %s", res.Status)
	}
}

func TestValidateSecrets_MultipleResults(t *testing.T) {
	c := &fakeValidateClient{
		data: map[string]map[string]interface{}{
			"kv/app/one": {"token": "abc"},
			"kv/app/two": {},
		},
	}
	results := ValidateSecrets(c, "kv", []string{"app/one", "app/two", "app/three"}, []string{"token"})
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Status != "ok" {
		t.Errorf("expected ok for app/one, got %s", results[0].Status)
	}
	if results[1].Status != "missing_keys" {
		t.Errorf("expected missing_keys for app/two, got %s", results[1].Status)
	}
	if results[2].Status != "not_found" {
		t.Errorf("expected not_found for app/three, got %s", results[2].Status)
	}
}
