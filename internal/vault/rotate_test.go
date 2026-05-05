package vault

import (
	"errors"
	"testing"
)

type fakeRotator struct {
	data    map[string]map[string]interface{}
	written map[string]map[string]interface{}
	readErr error
	writeErr error
}

func (f *fakeRotator) ReadSecret(_, path string) (map[string]interface{}, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	if d, ok := f.data[path]; ok {
		copy := make(map[string]interface{}, len(d))
		for k, v := range d {
			copy[k] = v
		}
		return copy, nil
	}
	return map[string]interface{}{}, nil
}

func (f *fakeRotator) WriteSecret(_, path string, data map[string]interface{}) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	if f.written == nil {
		f.written = make(map[string]map[string]interface{})
	}
	f.written[path] = data
	return nil
}

func TestRotateSecret_Success(t *testing.T) {
	r := &fakeRotator{
		data: map[string]map[string]interface{}{
			"app/db": {"password": "old-pass"},
		},
	}
	res := RotateSecret(r, "secret", "app/db", []string{"password"})
	if !res.Success {
		t.Fatalf("expected success, got error: %v", res.Error)
	}
	if r.written["app/db"]["password"] == "old-pass" {
		t.Error("expected password to be rotated")
	}
}

func TestRotateSecret_ReadError(t *testing.T) {
	r := &fakeRotator{readErr: errors.New("read failed")}
	res := RotateSecret(r, "secret", "app/db", nil)
	if res.Success {
		t.Fatal("expected failure")
	}
}

func TestRotateSecret_WriteError(t *testing.T) {
	r := &fakeRotator{
		data:     map[string]map[string]interface{}{"app/db": {"key": "val"}},
		writeErr: errors.New("write failed"),
	}
	res := RotateSecret(r, "secret", "app/db", nil)
	if res.Success {
		t.Fatal("expected failure")
	}
}

func TestRotateSecrets_Summary(t *testing.T) {
	r := &fakeRotator{
		data: map[string]map[string]interface{}{
			"a": {"k": "v"},
			"b": {"k": "v"},
		},
		writeErr: nil,
	}
	_, sum := RotateSecrets(r, "secret", []string{"a", "b", "c"}, nil)
	if sum.Total != 3 {
		t.Errorf("expected total 3, got %d", sum.Total)
	}
	// "c" has no keys so it fails
	if sum.Failed != 1 {
		t.Errorf("expected 1 failure, got %d", sum.Failed)
	}
	if sum.Rotated != 2 {
		t.Errorf("expected 2 rotated, got %d", sum.Rotated)
	}
}
