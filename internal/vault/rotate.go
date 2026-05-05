package vault

import (
	"fmt"
	"time"
)

// RotateResult holds the outcome of a single secret rotation.
type RotateResult struct {
	Path    string
	OldKey  string
	Success bool
	Error   error
}

// RotateSummary holds aggregate counts for a rotation run.
type RotateSummary struct {
	Total   int
	Rotated int
	Failed  int
}

// Rotator can read and write secrets.
type Rotator interface {
	ReadSecret(mount, path string) (map[string]interface{}, error)
	WriteSecret(mount, path string, data map[string]interface{}) error
}

// RotateSecret appends a timestamp-based version tag to every value in the
// secret at the given path, effectively forcing a new secret version.
func RotateSecret(r Rotator, mount, path string, keys []string) RotateResult {
	data, err := r.ReadSecret(mount, path)
	if err != nil {
		return RotateResult{Path: path, Success: false, Error: err}
	}

	targets := keys
	if len(targets) == 0 {
		for k := range data {
			targets = append(targets, k)
		}
	}

	if len(targets) == 0 {
		return RotateResult{Path: path, Success: false, Error: fmt.Errorf("no keys found in secret")}
	}

	stamp := fmt.Sprintf("rotated-%d", time.Now().Unix())
	for _, k := range targets {
		if _, ok := data[k]; ok {
			data[k] = stamp
		}
	}

	if err := r.WriteSecret(mount, path, data); err != nil {
		return RotateResult{Path: path, OldKey: targets[0], Success: false, Error: err}
	}

	return RotateResult{Path: path, OldKey: targets[0], Success: true}
}

// RotateSecrets rotates a list of paths and returns results plus a summary.
func RotateSecrets(r Rotator, mount string, paths []string, keys []string) ([]RotateResult, RotateSummary) {
	var results []RotateResult
	var summary RotateSummary

	for _, p := range paths {
		res := RotateSecret(r, mount, p, keys)
		results = append(results, res)
		summary.Total++
		if res.Success {
			summary.Rotated++
		} else {
			summary.Failed++
		}
	}

	return results, summary
}
