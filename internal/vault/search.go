package vault

import (
	"strings"
)

// SearchResult holds a matching secret path and the keys that matched.
type SearchResult struct {
	Path string
	MatchedKeys []string
}

// SearchSecrets searches all secrets under the given mount for keys or values
// that contain the query string (case-insensitive).
func SearchSecrets(c *Client, mount, query string) ([]SearchResult, error) {
	paths, err := ListSecrets(c, mount, "")
	if err != nil {
		return nil, err
	}

	q := strings.ToLower(query)
	var results []SearchResult

	for _, path := range paths {
		secret, err := ReadSecret(c, mount, path)
		if err != nil {
			continue
		}

		var matched []string
		for k, v := range secret {
			key := strings.ToLower(k)
			val := strings.ToLower(toString(v))
			if strings.Contains(key, q) || strings.Contains(val, q) {
				matched = append(matched, k)
			}
		}

		if len(matched) > 0 {
			results = append(results, SearchResult{
				Path:        path,
				MatchedKeys: matched,
			})
		}
	}

	return results, nil
}
