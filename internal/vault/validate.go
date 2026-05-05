package vault

import (
	"fmt"
	"strings"
)

// ValidationResult holds the outcome of validating a single secret path.
type ValidationResult struct {
	Path    string
	Missing []string
	Status  string // "ok", "missing_keys", "not_found", "error"
	Err     error
}

// SecretReader is satisfied by Client.
type SecretReader interface {
	ReadSecret(mount, path string) (map[string]interface{}, error)
}

// ValidateSecret checks that a secret at the given path contains all required keys.
func ValidateSecret(c SecretReader, mount, path string, requiredKeys []string) ValidationResult {
	data, err := c.ReadSecret(mount, path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return ValidationResult{Path: path, Status: "not_found", Err: err}
		}
		return ValidationResult{Path: path, Status: "error", Err: err}
	}

	var missing []string
	for _, k := range requiredKeys {
		if _, ok := data[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return ValidationResult{
			Path:    path,
			Missing: missing,
			Status:  "missing_keys",
			Err:     fmt.Errorf("missing keys: %s", strings.Join(missing, ", ")),
		}
	}

	return ValidationResult{Path: path, Status: "ok"}
}

// ValidateSecrets runs ValidateSecret over a list of paths and returns all results.
func ValidateSecrets(c SecretReader, mount string, paths []string, requiredKeys []string) []ValidationResult {
	results := make([]ValidationResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, ValidateSecret(c, mount, p, requiredKeys))
	}
	return results
}
