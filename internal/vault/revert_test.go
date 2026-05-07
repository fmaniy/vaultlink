package vault

import (
	"errors"
	"testing"
)

type fakeRevertLogical struct {
	readFn  func(path string) (*fakeSecret, error)
	writeFn func(path string, data map[string]interface{}) (*fakeSecret, error)
}

func (f *fakeRevertLogical) Read(path string) (*fakeSecret, error)  { return f.readFn(path) }
func (f *fakeRevertLogical) Write(path string, data map[string]interface{}) (*fakeSecret, error) {
	return f.writeFn(path, data)
}

func newRevertClient(readFn func(string) (*fakeSecret, error), writeFn func(string, map[string]interface{}) (*fakeSecret, error)) *Client {
	return clientWithFakeLogical(&fakeLogicalAdapter{
		readFn:  func(p string) (map[string]interface{}, error) {
			s, err := readFn(p)
			if err != nil || s == nil {
				return nil, err
			}
			return s.Data, nil
		},
		writeFn: func(p string, d map[string]interface{}) (map[string]interface{}, error) {
			s, err := writeFn(p, d)
			if err != nil || s == nil {
				return nil, err
			}
			return s.Data, nil
		},
	})
}

func TestRevertSecret_Success(t *testing.T) {
	client := newRevertClient(
		func(p string) (*fakeSecret, error) {
			return &fakeSecret{Data: map[string]interface{}{"data": map[string]interface{}{"key": "old"}}}, nil
		},
		func(p string, d map[string]interface{}) (*fakeSecret, error) {
			return &fakeSecret{}, nil
		},
	)
	res := RevertSecret(client, "secret", "myapp/db", 2)
	if !res.Reverted {
		t.Fatalf("expected Reverted=true, got error: %v", res.Error)
	}
}

func TestRevertSecret_ReadError(t *testing.T) {
	client := newRevertClient(
		func(p string) (*fakeSecret, error) { return nil, errors.New("vault unreachable") },
		func(p string, d map[string]interface{}) (*fakeSecret, error) { return &fakeSecret{}, nil },
	)
	res := RevertSecret(client, "secret", "myapp/db", 1)
	if res.Reverted {
		t.Fatal("expected Reverted=false")
	}
	if res.Error == nil {
		t.Fatal("expected error")
	}
}

func TestRevertSecret_NotFound(t *testing.T) {
	client := newRevertClient(
		func(p string) (*fakeSecret, error) { return nil, nil },
		func(p string, d map[string]interface{}) (*fakeSecret, error) { return &fakeSecret{}, nil },
	)
	res := RevertSecret(client, "secret", "myapp/db", 3)
	if res.Reverted {
		t.Fatal("expected Reverted=false")
	}
	if res.Error == nil {
		t.Fatal("expected not-found error")
	}
}

func TestRevertSecret_WriteError(t *testing.T) {
	client := newRevertClient(
		func(p string) (*fakeSecret, error) {
			return &fakeSecret{Data: map[string]interface{}{"data": map[string]interface{}{"k": "v"}}}, nil
		},
		func(p string, d map[string]interface{}) (*fakeSecret, error) {
			return nil, errors.New("write denied")
		},
	)
	res := RevertSecret(client, "secret", "myapp/db", 1)
	if res.Reverted {
		t.Fatal("expected Reverted=false")
	}
	if res.Error == nil {
		t.Fatal("expected write error")
	}
}

func TestRevertSecrets_MultipleResults(t *testing.T) {
	client := newRevertClient(
		func(p string) (*fakeSecret, error) {
			return &fakeSecret{Data: map[string]interface{}{"data": map[string]interface{}{"x": "1"}}}, nil
		},
		func(p string, d map[string]interface{}) (*fakeSecret, error) { return &fakeSecret{}, nil },
	)
	results := RevertSecrets(client, "secret", []string{"a", "b", "c"}, 1)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if !r.Reverted {
			t.Errorf("expected %s to be reverted", r.Path)
		}
	}
}
