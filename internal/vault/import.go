package vault

import (
	"encoding/json"
	"fmt"
	"os"
)

// ImportResult holds the outcome of a single secret import operation.
type ImportResult struct {
	Path    string
	Status  string
	Err     error
}

// ImportSecrets reads secrets from a JSON file and writes them to Vault.
// The file format matches the output of ExportSecretsToFile.
func ImportSecrets(client *Client, mountPath, filePath string, overwrite bool) ([]ImportResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading import file: %w", err)
	}

	var secrets map[string]map[string]interface{}
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, fmt.Errorf("parsing import file: %w", err)
	}

	var results []ImportResult
	for path, kvData := range secrets {
		result := ImportResult{Path: path}

		if !overwrite {
			existing, err := ReadSecret(client, mountPath, path)
			if err == nil && existing != nil {
				result.Status = "skipped"
				results = append(results, result)
				continue
			}
		}

		if err := WriteSecret(client, mountPath, path, kvData); err != nil {
			result.Status = "error"
			result.Err = err
		} else {
			result.Status = "imported"
		}
		results = append(results, result)
	}

	return results, nil
}
