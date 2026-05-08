package vault

import (
	"errors"
	"testing"
)

func newFlattenClient(data map[string]map[string]interface{}) *Client {
	logical := &fakeLogical{
		readFn: func(path string) (map[string]interface{}, error) {
			// path arrives as "mount/data/secret"
			for k, v := range data {
				if contains(path, k) {
					return map[string]interface{}{"data": v}, nil
				}
			}
			return nil, errors.New("not found")
		},
	}
	return clientWithFakeLogical(logical)
}

func contains(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		(haystack == needle || len(haystack) > len(needle) &&
			(haystack[len(haystack)-len(needle):] == needle ||
				hystack[:len(needle)] == needle))
}

func TestFlattenSecrets_Success(t *testing.T) {
	client := newFlattenClient(map[string]map[string]interface{}{
		"myapp/prod": {"db_pass": "s3cr3t", "api_key": "abc123"},
	})

	results := FlattenSecrets(client, "secret", []string{"myapp/prod"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// results are sorted by key
	if results[0].Key != "api_key" {
		t.Errorf("expected first key api_key, got %s", results[0].Key)
	}
	if results[1].Key != "db_pass" {
		t.Errorf("expected second key db_pass, got %s", results[1].Key)
	}
	for _, r := range results {
		if r.Error != nil {
			t.Errorf("unexpected error: %v", r.Error)
		}
		if r.Path != "myapp/prod" {
			t.Errorf("unexpected path: %s", r.Path)
		}
	}
}

func TestFlattenSecrets_ReadError(t *testing.T) {
	client := newFlattenClient(map[string]map[string]interface{}{})

	results := FlattenSecrets(client, "secret", []string{"missing/path"})

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Error("expected error for missing path")
	}
}

func TestFlattenSecrets_MultiplePaths(t *testing.T) {
	client := newFlattenClient(map[string]map[string]interface{}{
		"app/a": {"x": "1"},
		"app/b": {"y": "2"},
	})

	results := FlattenSecrets(client, "secret", []string{"app/a", "app/b"})

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestFlattenKey(t *testing.T) {
	got := FlattenKey("myapp/prod", "db_pass")
	want := "myapp/prod.db_pass"
	if got != want {
		t.Errorf("FlattenKey = %q, want %q", got, want)
	}
}

func TestFlattenKey_TrailingSlash(t *testing.T) {
	got := FlattenKey("myapp/prod/", "token")
	want := "myapp/prod.token"
	if got != want {
		t.Errorf("FlattenKey = %q, want %q", got, want)
	}
}
