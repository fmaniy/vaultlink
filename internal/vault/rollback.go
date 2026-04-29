package vault

import (
	"fmt"
)

// SecretSnapshot holds the state of a secret at a point in time.
type SecretSnapshot struct {
	Path string
	Data map[string]interface{}
}

// RollbackClient defines the interface needed for rollback operations.
type RollbackClient interface {
	Reader
	Writer
}

// TakeSnapshot reads the current value of a secret and returns a snapshot.
func TakeSnapshot(client RollbackClient, mount, path string) (*SecretSnapshot, error) {
	data, err := ReadSecret(client, mount, path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: read %s/%s: %w", mount, path, err)
	}

	copy := make(map[string]interface{}, len(data))
	for k, v := range data {
		copy[k] = v
	}

	return &SecretSnapshot{
		Path: path,
		Data: copy,
	}, nil
}

// RestoreSnapshot writes a previously taken snapshot back to Vault.
func RestoreSnapshot(client RollbackClient, mount string, snap *SecretSnapshot) error {
	if snap == nil {
		return fmt.Errorf("rollback: snapshot is nil")
	}
	if err := WriteSecret(client, mount, snap.Path, snap.Data); err != nil {
		return fmt.Errorf("rollback: restore %s/%s: %w", mount, snap.Path, err)
	}
	return nil
}

// RollbackSecrets takes snapshots of all given paths, applies fn, and
// restores all snapshots if fn returns an error.
func RollbackSecrets(
	client RollbackClient,
	mount string,
	paths []string,
	fn func() error,
) error {
	snaps := make([]*SecretSnapshot, 0, len(paths))
	for _, p := range paths {
		snap, err := TakeSnapshot(client, mount, p)
		if err != nil {
			return fmt.Errorf("rollback: snapshot phase: %w", err)
		}
		snaps = append(snaps, snap)
	}

	if err := fn(); err != nil {
		for _, snap := range snaps {
			// best-effort restore; ignore individual errors
			_ = RestoreSnapshot(client, mount, snap)
		}
		return fmt.Errorf("rollback: operation failed, snapshots restored: %w", err)
	}

	return nil
}
