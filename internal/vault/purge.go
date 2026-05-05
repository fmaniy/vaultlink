package vault

import (
	"fmt"
	"strings"
)

// PurgeResult holds the outcome of a single secret purge operation.
type PurgeResult struct {
	Path    string
	Deleted bool
	Skipped bool
	Error   error
}

// PurgeSummary holds aggregate counts for a purge run.
type PurgeSummary struct {
	Deleted int
	Skipped int
	Errors  int
}

// secretDeleter is the interface required to delete a secret.
type secretDeleter interface {
	Delete(path string) error
	List(path string) ([]string, error)
}

// PurgeSecret deletes a single secret at the given path.
// If dryRun is true the deletion is simulated and the result is marked Skipped.
func PurgeSecret(client secretDeleter, mount, path string, dryRun bool) PurgeResult {
	full := strings.Trim(mount, "/") + "/metadata/" + strings.Trim(path, "/")
	if dryRun {
		return PurgeResult{Path: path, Skipped: true}
	}
	if err := client.Delete(full); err != nil {
		return PurgeResult{Path: path, Error: fmt.Errorf("delete %s: %w", full, err)}
	}
	return PurgeResult{Path: path, Deleted: true}
}

// PurgeSecrets deletes all secrets under the given paths.
func PurgeSecrets(client secretDeleter, mount string, paths []string, dryRun bool) ([]PurgeResult, PurgeSummary) {
	results := make([]PurgeResult, 0, len(paths))
	var summary PurgeSummary
	for _, p := range paths {
		r := PurgeSecret(client, mount, p, dryRun)
		results = append(results, r)
		switch {
		case r.Error != nil:
			summary.Errors++
		case r.Skipped:
			summary.Skipped++
		default:
			summary.Deleted++
		}
	}
	return results, summary
}
