package vault

import "fmt"

// TrimResult holds the outcome of trimming whitespace from a single secret's values.
type TrimResult struct {
	Path    string
	Trimmed int
	Skipped bool
	Err     error
}

// TrimSecretClient is the interface required by TrimSecret.
type TrimSecretClient interface {
	Reader
	Writer
}

// TrimSecret reads the secret at path, strips leading/trailing whitespace from
// every string value, and writes it back. If dryRun is true the write is
// skipped and only the count of fields that would change is reported.
func TrimSecret(client TrimSecretClient, mount, path string, dryRun bool) TrimResult {
	data, err := ReadSecret(client, mount, path)
	if err != nil {
		return TrimResult{Path: path, Err: fmt.Errorf("read: %w", err)}
	}
	if data == nil {
		return TrimResult{Path: path, Skipped: true}
	}

	trimmed := 0
	for k, v := range data {
		if s, ok := v.(string); ok {
			clean := trimSpace(s)
			if clean != s {
				data[k] = clean
				trimmed++
			}
		}
	}

	if trimmed > 0 && !dryRun {
		if err := WriteSecret(client, mount, path, data); err != nil {
			return TrimResult{Path: path, Err: fmt.Errorf("write: %w", err)}
		}
	}

	return TrimResult{Path: path, Trimmed: trimmed}
}

// TrimSecrets runs TrimSecret over a list of paths.
func TrimSecrets(client TrimSecretClient, mount string, paths []string, dryRun bool) []TrimResult {
	results := make([]TrimResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, TrimSecret(client, mount, p, dryRun))
	}
	return results
}

// trimSpace is a local helper so the package does not import "strings" solely
// for this purpose — it matches strings.TrimSpace behaviour.
func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
