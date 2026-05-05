package vault

import (
	"fmt"
	"strings"
)

// SanitizeResult holds the outcome of sanitizing a single secret.
type SanitizeResult struct {
	Path    string
	Removed []string
	Status  string
	Err     error
}

// SanitizeSecrets removes keys from secrets whose names match any of the given
// prefixes or suffixes. It is useful for stripping debug/temp fields before
// promoting secrets to production environments.
func SanitizeSecrets(c *Client, paths []string, prefixes, suffixes []string, dryRun bool) []SanitizeResult {
	results := make([]SanitizeResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, sanitizeOne(c, p, prefixes, suffixes, dryRun))
	}
	return results
}

func sanitizeOne(c *Client, path string, prefixes, suffixes []string, dryRun bool) SanitizeResult {
	res := SanitizeResult{Path: path}

	data, err := ReadSecret(c, path)
	if err != nil {
		res.Status = "error"
		res.Err = fmt.Errorf("read: %w", err)
		return res
	}

	cleaned := make(map[string]interface{}, len(data))
	for k, v := range data {
		if matchesPrefixOrSuffix(k, prefixes, suffixes) {
			res.Removed = append(res.Removed, k)
		} else {
			cleaned[k] = v
		}
	}

	if len(res.Removed) == 0 {
		res.Status = "clean"
		return res
	}

	if dryRun {
		res.Status = "dry-run"
		return res
	}

	if err := WriteSecret(c, path, cleaned); err != nil {
		res.Status = "error"
		res.Err = fmt.Errorf("write: %w", err)
		return res
	}

	res.Status = "sanitized"
	return res
}

func matchesPrefixOrSuffix(key string, prefixes, suffixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	for _, s := range suffixes {
		if strings.HasSuffix(key, s) {
			return true
		}
	}
	return false
}
