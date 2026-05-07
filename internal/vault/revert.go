package vault

import "fmt"

// RevertResult holds the outcome of reverting a single secret.
type RevertResult struct {
	Path    string
	Reverted bool
	Error   error
}

// RevertSecret restores a secret at path to a previous version by reading
// that version and writing it back as the current data.
func RevertSecret(client *Client, mount, path string, version int) RevertResult {
	versioned := fmt.Sprintf("%s/data/%s", mount, path)

	// Read the specific version
	secret, err := client.Logical().Read(fmt.Sprintf("%s?version=%d", versioned, version))
	if err != nil {
		return RevertResult{Path: path, Error: fmt.Errorf("read version %d: %w", version, err)}
	}
	if secret == nil || secret.Data == nil {
		return RevertResult{Path: path, Error: fmt.Errorf("version %d not found at %s", version, path)}
	}

	data, ok := secret.Data["data"]
	if !ok {
		return RevertResult{Path: path, Error: fmt.Errorf("no data field in version %d of %s", version, path)}
	}

	// Write it back as a new current version
	payload := map[string]interface{}{"data": data}
	_, err = client.Logical().Write(versioned, payload)
	if err != nil {
		return RevertResult{Path: path, Error: fmt.Errorf("write revert: %w", err)}
	}

	return RevertResult{Path: path, Reverted: true}
}

// RevertSecrets reverts each path in paths to the given version.
func RevertSecrets(client *Client, mount string, paths []string, version int) []RevertResult {
	results := make([]RevertResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, RevertSecret(client, mount, p, version))
	}
	return results
}
