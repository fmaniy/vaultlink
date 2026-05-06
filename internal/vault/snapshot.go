package vault

import (
	"fmt"
	"time"
)

// SnapshotEntry holds a point-in-time capture of a single secret path.
type SnapshotEntry struct {
	Path      string
	Data      map[string]interface{}
	CapturedAt time.Time
}

// SnapshotResult is the outcome of snapshotting one secret.
type SnapshotResult struct {
	Path  string
	OK    bool
	Error string
}

// Snapshotter is the interface required for snapshot operations.
type Snapshotter interface {
	Reader
	Lister
}

// SnapshotSecrets reads every secret under the given mount/prefix and
// returns a slice of SnapshotEntry values along with per-path results.
func SnapshotSecrets(client Snapshotter, mount, prefix string) ([]SnapshotEntry, []SnapshotResult, error) {
	paths, err := ListSecrets(client, mount, prefix)
	if err != nil {
		return nil, nil, fmt.Errorf("list %s/%s: %w", mount, prefix, err)
	}

	now := time.Now().UTC()
	var entries []SnapshotEntry
	var results []SnapshotResult

	for _, p := range paths {
		data, err := ReadSecret(client, mount, p)
		if err != nil {
			results = append(results, SnapshotResult{Path: p, OK: false, Error: err.Error()})
			continue
		}
		entries = append(entries, SnapshotEntry{Path: p, Data: data, CapturedAt: now})
		results = append(results, SnapshotResult{Path: p, OK: true})
	}

	return entries, results, nil
}
