package vault

import (
	"errors"
	"testing"
)

func compareClient(data map[string]map[string]interface{}, keys []string) *fakeLogical {
	return &fakeLogical{
		readFn: func(path string) (map[string]interface{}, error) {
			if v, ok := data[path]; ok {
				return v, nil
			}
			return nil, errors.New("not found")
		},
		listFn: func(path string) ([]string, error) {
			return keys, nil
		},
	}
}

func TestCompareSecrets_AllMatch(t *testing.T) {
	payload := map[string]interface{}{"token": "abc"}
	src := compareClient(map[string]map[string]interface{}{"secret/data/app": payload}, []string{"app"})
	dst := compareClient(map[string]map[string]interface{}{"secret/data/app": payload}, []string{"app"})

	results, err := CompareSecrets(src, dst, "secret", "secret", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "match" {
		t.Errorf("expected match, got %+v", results)
	}
}

func TestCompareSecrets_MissingInDst(t *testing.T) {
	payload := map[string]interface{}{"key": "val"}
	src := compareClient(map[string]map[string]interface{}{"secret/data/app": payload}, []string{"app"})
	dst := compareClient(map[string]map[string]interface{}{}, []string{})

	results, err := CompareSecrets(src, dst, "secret", "secret", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "missing_dst" {
		t.Errorf("expected missing_dst, got %+v", results)
	}
}

func TestCompareSecrets_MissingInSrc(t *testing.T) {
	payload := map[string]interface{}{"key": "val"}
	src := compareClient(map[string]map[string]interface{}{}, []string{})
	dst := compareClient(map[string]map[string]interface{}{"secret/data/app": payload}, []string{"app"})

	results, err := CompareSecrets(src, dst, "secret", "secret", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "missing_src" {
		t.Errorf("expected missing_src, got %+v", results)
	}
}

func TestCompareSecrets_Mismatch(t *testing.T) {
	src := compareClient(map[string]map[string]interface{}{"secret/data/app": {"key": "v1"}}, []string{"app"})
	dst := compareClient(map[string]map[string]interface{}{"secret/data/app": {"key": "v2"}}, []string{"app"})

	results, err := CompareSecrets(src, dst, "secret", "secret", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "mismatch" {
		t.Errorf("expected mismatch, got %+v", results)
	}
	if results[0].Details == "" {
		t.Error("expected non-empty details for mismatch")
	}
}
