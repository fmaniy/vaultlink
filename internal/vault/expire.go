package vault

import (
	"fmt"
	"time"
)

// ExpiryResult holds the outcome of an expiry check for a single secret.
type ExpiryResult struct {
	Path      string
	ExpiresAt time.Time
	Expired   bool
	Missing   bool
	Error     error
}

// expiryReader is satisfied by the vault logical client.
type expiryReader interface {
	Read(path string) (map[string]interface{}, error)
}

// CheckExpiry reads the __expires_at metadata key written by vaultlink
// conventions and reports whether each secret has expired.
func CheckExpiry(client expiryReader, mount string, paths []string, now time.Time) []ExpiryResult {
	results := make([]ExpiryResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, checkOne(client, mount, p, now))
	}
	return results
}

func checkOne(client expiryReader, mount, path string, now time.Time) ExpiryResult {
	full := fmt.Sprintf("%s/data/%s", mount, path)
	data, err := client.Read(full)
	if err != nil {
		return ExpiryResult{Path: path, Error: err}
	}
	if data == nil {
		return ExpiryResult{Path: path, Missing: true}
	}

	inner, _ := data["data"].(map[string]interface{})
	if inner == nil {
		return ExpiryResult{Path: path, Missing: true}
	}

	raw, ok := inner["__expires_at"]
	if !ok {
		// No expiry tag — treat as never-expiring.
		return ExpiryResult{Path: path}
	}

	ts, ok := raw.(string)
	if !ok {
		return ExpiryResult{Path: path, Error: fmt.Errorf("__expires_at is not a string")}
	}

	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ExpiryResult{Path: path, Error: fmt.Errorf("parse __expires_at: %w", err)}
	}

	return ExpiryResult{
		Path:      path,
		ExpiresAt: t,
		Expired:   now.After(t),
	}
}
