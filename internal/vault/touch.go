package vault

import (
	"fmt"
	"time"
)

// TouchResult holds the outcome of a touch operation on a single secret.
type TouchResult struct {
	Path    string
	Updated bool
	Error   error
}

// Logical is the interface used to read/write Vault logical paths.
type touchWriter interface {
	Read(path string) (map[string]interface{}, error)
	Write(path string, data map[string]interface{}) error
}

// TouchSecret reads a secret and re-writes it with an added metadata field
// "_touched_at" to record when it was last touched without altering real data.
func TouchSecret(client touchWriter, mount, path string) TouchResult {
	fullPath := fmt.Sprintf("%s/data/%s", mount, path)

	existing, err := client.Read(fullPath)
	if err != nil {
		return TouchResult{Path: path, Error: fmt.Errorf("read: %w", err)}
	}
	if existing == nil {
		return TouchResult{Path: path, Error: fmt.Errorf("secret not found: %s", path)}
	}

	data, _ := existing["data"].(map[string]interface{})
	if data == nil {
		data = map[string]interface{}{}
	}
	data["_touched_at"] = time.Now().UTC().Format(time.RFC3339)

	if err := client.Write(fullPath, map[string]interface{}{"data": data}); err != nil {
		return TouchResult{Path: path, Error: fmt.Errorf("write: %w", err)}
	}

	return TouchResult{Path: path, Updated: true}
}

// TouchSecrets applies TouchSecret to each path and returns all results.
func TouchSecrets(client touchWriter, mount string, paths []string) []TouchResult {
	results := make([]TouchResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, TouchSecret(client, mount, p))
	}
	return results
}
