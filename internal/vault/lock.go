package vault

import (
	"fmt"
	"time"
)

// LockResult represents the outcome of a lock or unlock operation on a single path.
type LockResult struct {
	Path    string
	Action  string // "lock" or "unlock"
	Success bool
	Error   error
}

// LockWriter is the interface required to write lock metadata secrets.
type LockWriter interface {
	WriteSecret(mount, path string, data map[string]interface{}) error
	ReadSecret(mount, path string) (map[string]interface{}, error)
}

const lockMetaKey = "__vaultlink_locked"

// LockSecret writes a lock marker to the secret at the given path.
func LockSecret(client LockWriter, mount, path, lockedBy string) LockResult {
	existing, err := client.ReadSecret(mount, path)
	if err != nil {
		return LockResult{Path: path, Action: "lock", Success: false, Error: fmt.Errorf("read failed: %w", err)}
	}
	if existing == nil {
		existing = map[string]interface{}{}
	}
	existing[lockMetaKey] = fmt.Sprintf("%s@%s", lockedBy, time.Now().UTC().Format(time.RFC3339))
	if err := client.WriteSecret(mount, path, existing); err != nil {
		return LockResult{Path: path, Action: "lock", Success: false, Error: fmt.Errorf("write failed: %w", err)}
	}
	return LockResult{Path: path, Action: "lock", Success: true}
}

// UnlockSecret removes the lock marker from the secret at the given path.
func UnlockSecret(client LockWriter, mount, path string) LockResult {
	existing, err := client.ReadSecret(mount, path)
	if err != nil {
		return LockResult{Path: path, Action: "unlock", Success: false, Error: fmt.Errorf("read failed: %w", err)}
	}
	if existing == nil {
		return LockResult{Path: path, Action: "unlock", Success: false, Error: fmt.Errorf("secret not found")}
	}
	delete(existing, lockMetaKey)
	if err := client.WriteSecret(mount, path, existing); err != nil {
		return LockResult{Path: path, Action: "unlock", Success: false, Error: fmt.Errorf("write failed: %w", err)}
	}
	return LockResult{Path: path, Action: "unlock", Success: true}
}

// LockSecrets applies LockSecret to a list of paths and returns all results.
func LockSecrets(client LockWriter, mount string, paths []string, lockedBy string) []LockResult {
	results := make([]LockResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, LockSecret(client, mount, p, lockedBy))
	}
	return results
}
