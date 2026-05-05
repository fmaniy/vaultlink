package vault

import "fmt"

// PatchResult holds the outcome of a single patch operation.
type PatchResult struct {
	Path    string
	Key     string
	OldVal  string
	NewVal  string
	Skipped bool
	Err     error
}

// PatchSecret updates a single key within an existing secret at path.
// If the key does not exist and createKey is false, the operation is skipped.
func PatchSecret(c *Client, path, key, value string, createKey bool) PatchResult {
	res := PatchResult{Path: path, Key: key, NewVal: value}

	existing, err := ReadSecret(c, path)
	if err != nil {
		res.Err = fmt.Errorf("read %s: %w", path, err)
		return res
	}

	if existing == nil {
		res.Err = fmt.Errorf("secret not found: %s", path)
		return res
	}

	if old, ok := existing[key]; ok {
		res.OldVal = fmt.Sprintf("%v", old)
	} else if !createKey {
		res.Skipped = true
		return res
	}

	existing[key] = value
	if err := WriteSecret(c, path, existing); err != nil {
		res.Err = fmt.Errorf("write %s: %w", path, err)
		return res
	}
	return res
}

// PatchSecrets applies PatchSecret across multiple paths.
func PatchSecrets(c *Client, paths []string, key, value string, createKey bool) []PatchResult {
	results := make([]PatchResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, PatchSecret(c, p, key, value, createKey))
	}
	return results
}
