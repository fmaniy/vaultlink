package vault

import (
	"fmt"
)

// RenameSecret copies a secret from srcPath to dstPath within the same client,
// then deletes the original secret at srcPath.
func RenameSecret(client *Client, mount, srcPath, dstPath string) error {
	if srcPath == dstPath {
		return fmt.Errorf("source and destination paths are identical: %q", srcPath)
	}

	data, err := ReadSecret(client, mount, srcPath)
	if err != nil {
		return fmt.Errorf("rename: read source %q: %w", srcPath, err)
	}

	if err := WriteSecret(client, mount, dstPath, data); err != nil {
		return fmt.Errorf("rename: write destination %q: %w", dstPath, err)
	}

	if err := DeleteSecret(client, mount, srcPath); err != nil {
		return fmt.Errorf("rename: delete source %q: %w", srcPath, err)
	}

	return nil
}

// RenameSecrets renames each entry in the paths map from its key (src) to its
// value (dst). It stops and returns on the first error.
func RenameSecrets(client *Client, mount string, paths map[string]string) error {
	for src, dst := range paths {
		if err := RenameSecret(client, mount, src, dst); err != nil {
			return err
		}
	}
	return nil
}
