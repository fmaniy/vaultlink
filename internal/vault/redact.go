package vault

import (
	"fmt"
	"strings"
)

// RedactResult holds the outcome of redacting a single secret.
type RedactResult struct {
	Path    string
	Redacted int
	Skipped  int
	Err     error
}

// RedactSecrets replaces the values of matching keys with a placeholder
// across all provided paths. Keys are matched case-insensitively against
// the supplied patterns (e.g. "password", "token", "secret").
func RedactSecrets(client SecretReadWriter, mount string, paths []string, patterns []string, placeholder string, dryRun bool) []RedactResult {
	results := make([]RedactResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, redactOne(client, mount, p, patterns, placeholder, dryRun))
	}
	return results
}

func redactOne(client SecretReadWriter, mount, path string, patterns []string, placeholder string, dryRun bool) RedactResult {
	res := RedactResult{Path: path}

	data, err := ReadSecret(client, mount, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		return res
	}

	updated := make(map[string]interface{}, len(data))
	for k, v := range data {
		if matchesPattern(k, patterns) {
			updated[k] = placeholder
			res.Redacted++
		} else {
			updated[k] = v
			res.Skipped++
		}
	}

	if res.Redacted == 0 || dryRun {
		return res
	}

	if err := WriteSecret(client, mount, path, updated); err != nil {
		res.Err = fmt.Errorf("write %s: %w", path, err)
	}
	return res
}

func matchesPattern(key string, patterns []string) bool {
	lower := strings.ToLower(key)
	for _, p := range patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
