package vault

import "fmt"

// MergeResult holds the outcome of merging a single secret path.
type MergeResult struct {
	Path      string
	Status    string // "merged", "skipped", "error"
	Error     error
	Conflicts []string // keys that conflicted
}

// MergeOptions controls merge behaviour.
type MergeOptions struct {
	// Overwrite existing keys in dst when a conflict is detected.
	Overwrite bool
	// DryRun reports what would happen without writing.
	DryRun bool
}

// secretReadWriter can read and write secrets.
type secretReadWriter interface {
	ReadSecret(mount, path string) (map[string]interface{}, error)
	WriteSecret(mount, path string, data map[string]interface{}) error
}

// MergeSecret merges the KV data from srcMount/srcPath into dstMount/dstPath.
// Keys present in dst but absent in src are preserved.
func MergeSecret(c secretReadWriter, srcMount, srcPath, dstMount, dstPath string, opts MergeOptions) MergeResult {
	result := MergeResult{Path: dstPath}

	srcData, err := c.ReadSecret(srcMount, srcPath)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Errorf("read src: %w", err)
		return result
	}

	dstData, err := c.ReadSecret(dstMount, dstPath)
	if err != nil {
		dstData = map[string]interface{}{}
	}

	merged := make(map[string]interface{}, len(dstData))
	for k, v := range dstData {
		merged[k] = v
	}

	for k, v := range srcData {
		if existing, exists := merged[k]; exists {
			if fmt.Sprintf("%v", existing) != fmt.Sprintf("%v", v) {
				result.Conflicts = append(result.Conflicts, k)
				if !opts.Overwrite {
					continue
				}
			}
		}
		merged[k] = v
	}

	if opts.DryRun {
		result.Status = "skipped"
		return result
	}

	if err := c.WriteSecret(dstMount, dstPath, merged); err != nil {
		result.Status = "error"
		result.Error = fmt.Errorf("write dst: %w", err)
		return result
	}

	result.Status = "merged"
	return result
}

// MergeSecrets merges a slice of paths from src into dst.
func MergeSecrets(c secretReadWriter, srcMount string, paths []string, dstMount string, opts MergeOptions) []MergeResult {
	results := make([]MergeResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, MergeSecret(c, srcMount, p, dstMount, p, opts))
	}
	return results
}
