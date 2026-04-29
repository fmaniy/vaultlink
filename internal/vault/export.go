package vault

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ExportResult holds the result of exporting a single secret path.
type ExportResult struct {
	Path    string
	Success bool
	Error   string
}

// ExportSecrets reads all secrets under the given paths and writes them as JSON to w.
// The output format is a map of path -> key/value pairs.
func ExportSecrets(client *Client, mount string, paths []string, w io.Writer) ([]ExportResult, error) {
	output := make(map[string]map[string]string, len(paths))
	results := make([]ExportResult, 0, len(paths))

	for _, path := range paths {
		secret, err := ReadSecret(client, mount, path)
		if err != nil {
			results = append(results, ExportResult{Path: path, Success: false, Error: err.Error()})
			continue
		}
		output[path] = secret
		results = append(results, ExportResult{Path: path, Success: true})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		return results, fmt.Errorf("encoding export output: %w", err)
	}

	return results, nil
}

// ExportSecretsToFile is a convenience wrapper that writes to a named file.
func ExportSecretsToFile(client *Client, mount string, paths []string, filename string) ([]ExportResult, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("creating export file %q: %w", filename, err)
	}
	defer f.Close()
	return ExportSecrets(client, mount, paths, f)
}
