package vault

import (
	"fmt"
	"sort"
	"strings"
)

// FlattenResult holds the outcome of flattening a single secret path.
type FlattenResult struct {
	Path    string
	Key     string
	Value   string
	Error   error
}

// FlattenSecrets reads all secrets under the given paths and returns a flat
// list of path+key+value triples, suitable for export or inspection.
func FlattenSecrets(client *Client, mount string, paths []string) []FlattenResult {
	var results []FlattenResult

	for _, path := range paths {
		secret, err := ReadSecret(client, mount, path)
		if err != nil {
			results = append(results, FlattenResult{
				Path:  path,
				Error: err,
			})
			continue
		}

		keys := make([]string, 0, len(secret))
		for k := range secret {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := fmt.Sprintf("%v", secret[k])
			results = append(results, FlattenResult{
				Path:  path,
				Key:   k,
				Value: v,
			})
		}
	}

	return results
}

// FlattenKey returns a dot-joined representation of path and key,
// e.g. "myapp/prod" + "db_pass" => "myapp/prod.db_pass".
func FlattenKey(path, key string) string {
	path = strings.TrimSuffix(path, "/")
	return path + "." + key
}
