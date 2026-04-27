package vault

import (
	"fmt"
)

// CopySecret reads a secret from the source client and writes it to the destination client.
// It returns an error if the read or write fails.
func CopySecret(src, dst *Client, srcPath, dstPath string) error {
	data, err := ReadSecret(src, srcPath)
	if err != nil {
		return fmt.Errorf("copy: read from %q: %w", srcPath, err)
	}
	if err := WriteSecret(dst, dstPath, data); err != nil {
		return fmt.Errorf("copy: write to %q: %w", dstPath, err)
	}
	return nil
}

// CopySecrets copies all secrets found under srcPrefix on src to dstPrefix on dst.
// Paths are discovered via ListSecrets. Each leaf secret is copied individually.
func CopySecrets(src, dst *Client, srcMount, dstMount, srcPrefix, dstPrefix string) (int, error) {
	paths, err := ListSecrets(src, srcMount, srcPrefix)
	if err != nil {
		return 0, fmt.Errorf("copy: list secrets under %q: %w", srcPrefix, err)
	}

	copied := 0
	for _, rel := range paths {
		srcPath := srcMount + "/data/" + srcPrefix + rel
		dstPath := dstMount + "/data/" + dstPrefix + rel
		if err := CopySecret(src, dst, srcPath, dstPath); err != nil {
			return copied, err
		}
		copied++
	}
	return copied, nil
}
