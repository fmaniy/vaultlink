package vault

import "fmt"

// DedupeResult captures the outcome of a deduplication check for a single secret.
type DedupeResult struct {
	Path      string
	Duplicate bool
	MatchPath string // path of the first seen secret with identical data
	Err       error
}

// DedupeSecrets scans the given paths and identifies secrets whose entire data
// payload is identical to another secret in the list. The first occurrence is
// kept; subsequent identical secrets are flagged as duplicates.
func DedupeSecrets(client SecretReader, paths []string) []DedupeResult {
	type entry struct {
		path string
		data map[string]interface{}
	}

	seen := make([]entry, 0, len(paths))
	results := make([]DedupeResult, 0, len(paths))

	for _, p := range paths {
		data, err := ReadSecret(client, p)
		if err != nil {
			results = append(results, DedupeResult{Path: p, Err: fmt.Errorf("read: %w", err)})
			continue
		}

		matchPath := ""
		for _, e := range seen {
			if mapsEqual(e.data, data) {
				matchPath = e.path
				break
			}
		}

		if matchPath != "" {
			results = append(results, DedupeResult{Path: p, Duplicate: true, MatchPath: matchPath})
		} else {
			seen = append(seen, entry{path: p, data: data})
			results = append(results, DedupeResult{Path: p, Duplicate: false})
		}
	}

	return results
}

// mapsEqual performs a shallow equality check between two secret data maps.
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if fmt.Sprintf("%v", va) != fmt.Sprintf("%v", vb) {
			return false
		}
	}
	return true
}
