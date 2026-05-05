package vault

import (
	"errors"
	"testing"
	"time"
)

type fakeExpiryClient struct {
	data map[string]map[string]interface{}
	err  error
}

func (f *fakeExpiryClient) Read(path string) (map[string]interface{}, error) {
	if f.err != nil {
		return nil, f.err
	}
	v, ok := f.data[path]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func secret(kv map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"data": kv}
}

var refTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func TestCheckExpiry_NotExpired(t *testing.T) {
	client := &fakeExpiryClient{
		data: map[string]map[string]interface{}{
			"secret/data/myapp/db": secret(map[string]interface{}{
				"password":    "s3cr3t",
				"__expires_at": "2025-01-01T00:00:00Z",
			}),
		},
	}
	res := CheckExpiry(client, "secret", []string{"myapp/db"}, refTime)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if res[0].Expired {
		t.Error("expected not expired")
	}
	if res[0].Error != nil {
		t.Errorf("unexpected error: %v", res[0].Error)
	}
}

func TestCheckExpiry_Expired(t *testing.T) {
	client := &fakeExpiryClient{
		data: map[string]map[string]interface{}{
			"secret/data/myapp/token": secret(map[string]interface{}{
				"value":       "abc",
				"__expires_at": "2023-01-01T00:00:00Z",
			}),
		},
	}
	res := CheckExpiry(client, "secret", []string{"myapp/token"}, refTime)
	if !res[0].Expired {
		t.Error("expected expired")
	}
}

func TestCheckExpiry_NoTag(t *testing.T) {
	client := &fakeExpiryClient{
		data: map[string]map[string]interface{}{
			"secret/data/myapp/cfg": secret(map[string]interface{}{"key": "val"}),
		},
	}
	res := CheckExpiry(client, "secret", []string{"myapp/cfg"}, refTime)
	if res[0].Expired || res[0].Error != nil {
		t.Errorf("expected clean result, got expired=%v err=%v", res[0].Expired, res[0].Error)
	}
}

func TestCheckExpiry_Missing(t *testing.T) {
	client := &fakeExpiryClient{data: map[string]map[string]interface{}{}}
	res := CheckExpiry(client, "secret", []string{"myapp/gone"}, refTime)
	if !res[0].Missing {
		t.Error("expected missing")
	}
}

func TestCheckExpiry_ReadError(t *testing.T) {
	client := &fakeExpiryClient{err: errors.New("vault unavailable")}
	res := CheckExpiry(client, "secret", []string{"myapp/x"}, refTime)
	if res[0].Error == nil {
		t.Error("expected error")
	}
}

func TestCheckExpiry_BadTimestamp(t *testing.T) {
	client := &fakeExpiryClient{
		data: map[string]map[string]interface{}{
			"secret/data/myapp/bad": secret(map[string]interface{}{"__expires_at": "not-a-date"}),
		},
	}
	res := CheckExpiry(client, "secret", []string{"myapp/bad"}, refTime)
	if res[0].Error == nil {
		t.Error("expected parse error")
	}
}
