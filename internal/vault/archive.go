package vault

import (
	"fmt"
	"time"
)

// ArchiveResult holds the outcome of a single archive operation.
type ArchiveResult struct {
	Path      string
	ArchivedAt time.Time
	Err       error
}

// Archiver defines the interface required to archive secrets.
type Archiver interface {
	Reader
	Writer
}

// ArchiveSecret marks a secret as archived by writing an _archived_at metadata
// key into the secret's data. The original data is preserved.
func ArchiveSecret(client Archiver, mount, path string) ArchiveResult {
	result := ArchiveResult{Path: path}

	data, err := ReadSecret(client, mount, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}
	if data == nil {
		result.Err = fmt.Errorf("secret not found: %s", path)
		return result
	}

	if _, already := data["_archived_at"]; already {
		result.Err = fmt.Errorf("secret already archived: %s", path)
		return result
	}

	now := time.Now().UTC()
	data["_archived_at"] = now.Format(time.RFC3339)

	if err := WriteSecret(client, mount, path, data); err != nil {
		result.Err = fmt.Errorf("write %s: %w", path, err)
		return result
	}

	result.ArchivedAt = now
	return result
}

// UnarchiveSecret removes the _archived_at metadata key from a secret.
func UnarchiveSecret(client Archiver, mount, path string) ArchiveResult {
	result := ArchiveResult{Path: path}

	data, err := ReadSecret(client, mount, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}
	if data == nil {
		result.Err = fmt.Errorf("secret not found: %s", path)
		return result
	}

	if _, exists := data["_archived_at"]; !exists {
		result.Err = fmt.Errorf("secret is not archived: %s", path)
		return result
	}

	delete(data, "_archived_at")

	if err := WriteSecret(client, mount, path, data); err != nil {
		result.Err = fmt.Errorf("write %s: %w", path, err)
		return result
	}

	return result
}

// ArchiveSecrets runs ArchiveSecret over a list of paths.
func ArchiveSecrets(client Archiver, mount string, paths []string) []ArchiveResult {
	results := make([]ArchiveResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, ArchiveSecret(client, mount, p))
	}
	return results
}
