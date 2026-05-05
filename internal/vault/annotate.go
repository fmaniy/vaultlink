package vault

import (
	"fmt"
	"strings"
)

// AnnotateResult holds the outcome of a single annotate operation.
type AnnotateResult struct {
	Path    string
	Key     string
	Value   string
	Updated bool
	Err     error
}

// AnnotateSecret sets a metadata annotation (key=value) on a secret by storing
// it under a reserved "_annotations" map within the secret data.
func AnnotateSecret(c *Client, path, key, value string) AnnotateResult {
	result := AnnotateResult{Path: path, Key: key, Value: value}

	data, err := ReadSecret(c, path)
	if err != nil {
		result.Err = fmt.Errorf("read %s: %w", path, err)
		return result
	}

	annotations := extractAnnotations(data)
	annotations[key] = value
	data["_annotations"] = flattenAnnotations(annotations)

	if err := WriteSecret(c, path, data); err != nil {
		result.Err = fmt.Errorf("write %s: %w", path, err)
		return result
	}

	result.Updated = true
	return result
}

// AnnotateSecrets applies the same annotation to multiple paths.
func AnnotateSecrets(c *Client, paths []string, key, value string) []AnnotateResult {
	results := make([]AnnotateResult, 0, len(paths))
	for _, p := range paths {
		results = append(results, AnnotateSecret(c, p, key, value))
	}
	return results
}

func extractAnnotations(data map[string]interface{}) map[string]string {
	annotations := map[string]string{}
	raw, ok := data["_annotations"]
	if !ok {
		return annotations
	}
	str, ok := raw.(string)
	if !ok {
		return annotations
	}
	for _, pair := range strings.Split(str, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			annotations[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return annotations
}

func flattenAnnotations(annotations map[string]string) string {
	pairs := make([]string, 0, len(annotations))
	for k, v := range annotations {
		pairs = append(pairs, k+"="+v)
	}
	return strings.Join(pairs, ",")
}
