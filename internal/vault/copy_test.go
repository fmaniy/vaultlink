package vault

import (
	"errors"
	"testing"
)

// stubLogicalCopy wraps fakeLogical to allow selective read/write error injection.
type stubLogicalCopy struct {
	readData  map[string]interface{}
	written   map[string]map[string]interface{}
	readErr   error
	writeErr  error
}

func (s *stubLogicalCopy) Read(path string) (*fakeSecret, error) {
	if s.readErr != nil {
		return nil, s.readErr
	}
	if d, ok := s.readData[path]; ok {
		return &fakeSecret{Data: map[string]interface{}{"data": d}}, nil
	}
	return nil, nil
}

func (s *stubLogicalCopy) Write(path string, data map[string]interface{}) (*fakeSecret, error) {
	if s.writeErr != nil {
		return nil, s.writeErr
	}
	if s.written == nil {
		s.written = make(map[string]map[string]interface{})
	}
	s.written[path] = data
	return &fakeSecret{}, nil
}

func TestCopySecret_Success(t *testing.T) {
	payload := map[string]interface{}{"password": "s3cr3t"}
	stub := &stubLogicalCopy{
		readData: map[string]interface{}{"secret/data/app/db": payload},
	}
	src := clientWithFakeLogical(stub)
	dst := clientWithFakeLogical(stub)

	if err := CopySecret(src, dst, "secret/data/app/db", "secret/data/app/db"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.written["secret/data/app/db"] == nil {
		t.Error("expected secret to be written to destination")
	}
}

func TestCopySecret_ReadError(t *testing.T) {
	stub := &stubLogicalCopy{readErr: errors.New("permission denied")}
	src := clientWithFakeLogical(stub)
	dst := clientWithFakeLogical(stub)

	err := CopySecret(src, dst, "secret/data/app/db", "secret/data/app/db")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCopySecret_WriteError(t *testing.T) {
	payload := map[string]interface{}{"key": "val"}
	stub := &stubLogicalCopy{
		readData: map[string]interface{}{"secret/data/app/key": payload},
		writeErr: errors.New("storage error"),
	}
	src := clientWithFakeLogical(stub)
	dst := clientWithFakeLogical(stub)

	err := CopySecret(src, dst, "secret/data/app/key", "secret/data/app/key")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
