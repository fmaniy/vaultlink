package vault

import "fmt"

// PromoteResult holds the outcome of a single secret promotion.
type PromoteResult struct {
	Path    string
	Success bool
	Error   error
}

// PromoteSecret copies a secret from src to dst, optionally overwriting.
func PromoteSecret(src, dst Client, path string, overwrite bool) PromoteResult {
	if !overwrite {
		existing, err := dst.ReadSecret(path)
		if err == nil && existing != nil {
			return PromoteResult{
				Path:    path,
				Success: false,
				Error:   fmt.Errorf("secret already exists at %s (use --overwrite to replace)", path),
			}
		}
	}

	data, err := src.ReadSecret(path)
	if err != nil {
		return PromoteResult{Path: path, Success: false, Error: fmt.Errorf("read: %w", err)}
	}

	if err := dst.WriteSecret(path, data); err != nil {
		return PromoteResult{Path: path, Success: false, Error: fmt.Errorf("write: %w", err)}
	}

	return PromoteResult{Path: path, Success: true}
}

// PromoteSecrets promotes a slice of paths from src to dst.
func PromoteSecrets(src, dst Client, paths []string, overwrite bool) []PromoteResult {
	results := make([]PromoteResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, PromoteSecret(src, dst, p, overwrite))
	}
	return results
}
