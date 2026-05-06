package vault

import "fmt"

// TransferResult holds the outcome of a single secret transfer.
type TransferResult struct {
	Path    string
	Skipped bool
	Error   error
}

// TransferOptions controls transfer behaviour.
type TransferOptions struct {
	Overwrite bool
	DryRun    bool
}

// TransferSecret reads a secret from src at srcPath and writes it to dst at
// dstPath. If Overwrite is false and the destination already exists the
// transfer is skipped. If DryRun is true no write is performed.
func TransferSecret(src, dst Logical, srcPath, dstPath string, opts TransferOptions) TransferResult {
	res := TransferResult{Path: srcPath}

	data, err := ReadSecret(src, srcPath)
	if err != nil {
		res.Error = fmt.Errorf("read %s: %w", srcPath, err)
		return res
	}

	if !opts.Overwrite {
		existing, _ := ReadSecret(dst, dstPath)
		if existing != nil {
			res.Skipped = true
			return res
		}
	}

	if opts.DryRun {
		return res
	}

	if err := WriteSecret(dst, dstPath, data); err != nil {
		res.Error = fmt.Errorf("write %s: %w", dstPath, err)
	}
	return res
}

// TransferSecrets transfers multiple secrets from src to dst, rewriting the
// path prefix from srcMount to dstMount.
func TransferSecrets(src, dst Logical, srcMount, dstMount string, paths []string, opts TransferOptions) []TransferResult {
	results := make([]TransferResult, 0, len(paths))
	for _, p := range paths {
		dstPath := dstMount + "/" + trimMount(srcMount, p)
		results = append(results, TransferSecret(src, dst, p, dstPath, opts))
	}
	return results
}

// trimMount strips the mount prefix from a full secret path.
func trimMount(mount, path string) string {
	prefix := mount + "/"
	if len(path) > len(prefix) && path[:len(prefix)] == prefix {
		return path[len(prefix):]
	}
	return path
}
