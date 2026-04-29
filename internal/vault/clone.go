package vault

import "fmt"

// CloneResult holds the outcome of a single secret clone operation.
type CloneResult struct {
	SourcePath string
	DestPath   string
	Status     string // "cloned", "skipped", "error"
	Err        error
}

// CloneSecret copies a secret from srcPath to dstPath, optionally overwriting.
func CloneSecret(src, dst Logical, srcMount, dstMount, srcPath, dstPath string, overwrite bool) CloneResult {
	result := CloneResult{SourcePath: srcPath, DestPath: dstPath}

	data, err := ReadSecret(src, srcMount, srcPath)
	if err != nil {
		result.Status = "error"
		result.Err = fmt.Errorf("read %s: %w", srcPath, err)
		return result
	}

	if !overwrite {
		existing, _ := ReadSecret(dst, dstMount, dstPath)
		if existing != nil {
			result.Status = "skipped"
			return result
		}
	}

	if err := WriteSecret(dst, dstMount, dstPath, data); err != nil {
		result.Status = "error"
		result.Err = fmt.Errorf("write %s: %w", dstPath, err)
		return result
	}

	result.Status = "cloned"
	return result
}

// CloneSecrets clones a list of paths from one mount/env to another.
func CloneSecrets(src, dst Logical, srcMount, dstMount string, paths []string, overwrite bool) []CloneResult {
	results := make([]CloneResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, CloneSecret(src, dst, srcMount, dstMount, p, p, overwrite))
	}
	return results
}
