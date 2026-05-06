package vault

import (
	"errors"
	"testing"
)

func newTransferClient(data map[string]map[string]interface{}) *fakeLogical {
	return &fakeLogical{data: data}
}

func TestTransferSecret_Success(t *testing.T) {
	src := newTransferClient(map[string]map[string]interface{}{
		"secret/data/app/db": {"password": "s3cr3t"},
	})
	dst := newTransferClient(map[string]map[string]interface{}{})

	result := TransferSecret(src, dst, "secret/data/app/db", "backup/data/app/db", TransferOptions{})

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Skipped {
		t.Fatal("expected not skipped")
	}
	if dst.data["backup/data/app/db"] == nil {
		t.Fatal("expected secret written to dst")
	}
}

func TestTransferSecret_SkippedWhenExists(t *testing.T) {
	src := newTransferClient(map[string]map[string]interface{}{
		"secret/data/app/db": {"password": "s3cr3t"},
	})
	dst := newTransferClient(map[string]map[string]interface{}{
		"backup/data/app/db": {"password": "old"},
	})

	result := TransferSecret(src, dst, "secret/data/app/db", "backup/data/app/db", TransferOptions{Overwrite: false})

	if !result.Skipped {
		t.Fatal("expected skipped")
	}
}

func TestTransferSecret_OverwriteReplaces(t *testing.T) {
	src := newTransferClient(map[string]map[string]interface{}{
		"secret/data/app/db": {"password": "new"},
	})
	dst := newTransferClient(map[string]map[string]interface{}{
		"backup/data/app/db": {"password": "old"},
	})

	result := TransferSecret(src, dst, "secret/data/app/db", "backup/data/app/db", TransferOptions{Overwrite: true})

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if result.Skipped {
		t.Fatal("expected not skipped")
	}
	if dst.data["backup/data/app/db"]["password"] != "new" {
		t.Fatal("expected overwritten value")
	}
}

func TestTransferSecret_ReadError(t *testing.T) {
	src := &fakeLogical{err: errors.New("vault unreachable")}
	dst := newTransferClient(map[string]map[string]interface{}{})

	result := TransferSecret(src, dst, "secret/data/app/db", "backup/data/app/db", TransferOptions{})

	if result.Error == nil {
		t.Fatal("expected error")
	}
}

func TestTransferSecret_DryRun(t *testing.T) {
	src := newTransferClient(map[string]map[string]interface{}{
		"secret/data/app/db": {"password": "s3cr3t"},
	})
	dst := newTransferClient(map[string]map[string]interface{}{})

	result := TransferSecret(src, dst, "secret/data/app/db", "backup/data/app/db", TransferOptions{DryRun: true})

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}
	if len(dst.data) != 0 {
		t.Fatal("expected no write during dry-run")
	}
}
