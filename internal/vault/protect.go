package vault

import "fmt"

const protectMetaKey = "__protected"

// ProtectResult holds the outcome of a protect/unprotect operation.
type ProtectResult struct {
	Path      string
	Protected bool
	Skipped   bool
	Error     error
}

type protectReader interface {
	ReadSecret(mount, path string) (map[string]interface{}, error)
	WriteSecret(mount, path string, data map[string]interface{}) error
}

// ProtectSecret marks a secret as protected by setting a metadata flag.
// Protected secrets will be skipped by write operations that respect this flag.
func ProtectSecret(c protectReader, mount, path string) ProtectResult {
	data, err := c.ReadSecret(mount, path)
	if err != nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("read: %w", err)}
	}
	if data == nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("secret not found: %s", path)}
	}
	if data[protectMetaKey] == "true" {
		return ProtectResult{Path: path, Protected: true, Skipped: true}
	}
	data[protectMetaKey] = "true"
	if err := c.WriteSecret(mount, path, data); err != nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("write: %w", err)}
	}
	return ProtectResult{Path: path, Protected: true}
}

// UnprotectSecret removes the protection flag from a secret.
func UnprotectSecret(c protectReader, mount, path string) ProtectResult {
	data, err := c.ReadSecret(mount, path)
	if err != nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("read: %w", err)}
	}
	if data == nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("secret not found: %s", path)}
	}
	if data[protectMetaKey] != "true" {
		return ProtectResult{Path: path, Protected: false, Skipped: true}
	}
	delete(data, protectMetaKey)
	if err := c.WriteSecret(mount, path, data); err != nil {
		return ProtectResult{Path: path, Error: fmt.Errorf("write: %w", err)}
	}
	return ProtectResult{Path: path, Protected: false}
}

// ProtectSecrets applies ProtectSecret across multiple paths.
func ProtectSecrets(c protectReader, mount string, paths []string, unprotect bool) []ProtectResult {
	results := make([]ProtectResult, 0, len(paths))
	for _, p := range paths {
		var r ProtectResult
		if unprotect {
			r = UnprotectSecret(c, mount, p)
		} else {
			r = ProtectSecret(c, mount, p)
		}
		results = append(results, r)
	}
	return results
}
