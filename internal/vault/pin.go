package vault

import (
	"fmt"
	"time"
)

const pinVersionKey = "__pinned_version"
const pinTimestampKey = "__pinned_at"

// PinResult holds the outcome of a single pin or unpin operation.
type PinResult struct {
	Path    string
	Version string
	Pinned  bool
	Error   error
}

// PinSecret records the current version of a secret by writing pin metadata
// into the secret's data map. It reads the secret, stamps the version and
// timestamp, then writes it back.
func PinSecret(client *Client, mount, path, version string) PinResult {
	result := PinResult{Path: path, Version: version}

	data, err := ReadSecret(client, mount, path)
	if err != nil {
		result.Error = fmt.Errorf("read: %w", err)
		return result
	}
	if data == nil {
		result.Error = fmt.Errorf("secret not found: %s", path)
		return result
	}

	data[pinVersionKey] = version
	data[pinTimestampKey] = time.Now().UTC().Format(time.RFC3339)

	if err := WriteSecret(client, mount, path, data); err != nil {
		result.Error = fmt.Errorf("write: %w", err)
		return result
	}

	result.Pinned = true
	return result
}

// UnpinSecret removes pin metadata from a secret.
func UnpinSecret(client *Client, mount, path string) PinResult {
	result := PinResult{Path: path}

	data, err := ReadSecret(client, mount, path)
	if err != nil {
		result.Error = fmt.Errorf("read: %w", err)
		return result
	}
	if data == nil {
		result.Error = fmt.Errorf("secret not found: %s", path)
		return result
	}

	delete(data, pinVersionKey)
	delete(data, pinTimestampKey)

	if err := WriteSecret(client, mount, path, data); err != nil {
		result.Error = fmt.Errorf("write: %w", err)
		return result
	}

	result.Pinned = false
	return result
}

// PinSecrets applies PinSecret across multiple paths.
func PinSecrets(client *Client, mount string, paths []string, version string) []PinResult {
	results := make([]PinResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, PinSecret(client, mount, p, version))
	}
	return results
}
