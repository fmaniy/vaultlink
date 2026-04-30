package vault

import (
	"fmt"
	"strings"
)

const tagMetaKey = "vaultlink_tags"

// TagResult holds the outcome of a tag operation on a single secret path.
type TagResult struct {
	Path    string
	Tags    []string
	Success bool
	Err     error
}

// TagSecret adds the given tags to the secret at path by storing them in its metadata.
func TagSecret(client *Client, mount, path string, tags []string) TagResult {
	existing, err := ReadSecret(client, mount, path)
	if err != nil {
		return TagResult{Path: path, Err: fmt.Errorf("read: %w", err)}
	}

	current := ""
	if v, ok := existing[tagMetaKey]; ok {
		current, _ = v.(string)
	}

	merged := mergeTags(current, tags)
	existing[tagMetaKey] = strings.Join(merged, ",")

	if err := WriteSecret(client, mount, path, existing); err != nil {
		return TagResult{Path: path, Err: fmt.Errorf("write: %w", err)}
	}
	return TagResult{Path: path, Tags: merged, Success: true}
}

// UntagSecret removes the given tags from the secret at path.
func UntagSecret(client *Client, mount, path string, tags []string) TagResult {
	existing, err := ReadSecret(client, mount, path)
	if err != nil {
		return TagResult{Path: path, Err: fmt.Errorf("read: %w", err)}
	}

	current := ""
	if v, ok := existing[tagMetaKey]; ok {
		current, _ = v.(string)
	}

	remaining := removeTags(splitTags(current), tags)
	existing[tagMetaKey] = strings.Join(remaining, ",")

	if err := WriteSecret(client, mount, path, existing); err != nil {
		return TagResult{Path: path, Err: fmt.Errorf("write: %w", err)}
	}
	return TagResult{Path: path, Tags: remaining, Success: true}
}

// TagSecrets applies TagSecret across multiple paths.
func TagSecrets(client *Client, mount string, paths, tags []string) []TagResult {
	results := make([]TagResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, TagSecret(client, mount, p, tags))
	}
	return results
}

func splitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func mergeTags(current string, add []string) []string {
	seen := map[string]struct{}{}
	existing := splitTags(current)
	for _, t := range existing {
		seen[t] = struct{}{}
	}
	for _, t := range add {
		if _, ok := seen[t]; !ok {
			existing = append(existing, t)
			seen[t] = struct{}{}
		}
	}
	return existing
}

func removeTags(current, remove []string) []string {
	drop := map[string]struct{}{}
	for _, t := range remove {
		drop[t] = struct{}{}
	}
	out := current[:0]
	for _, t := range current {
		if _, skip := drop[t]; !skip {
			out = append(out, t)
		}
	}
	return out
}
