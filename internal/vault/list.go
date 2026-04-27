package vault

import (
	"fmt"
	"strings"
)

// ListSecrets returns all secret keys under the given path for a KV v2 mount.
// It performs a LIST operation against the metadata endpoint.
func (c *Client) ListSecrets(mount, path string) ([]string, error) {
	metaPath := buildMetaPath(mount, path)

	secret, err := c.logical.List(metaPath)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", metaPath, err)
	}
	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	raw, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	ifaces, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for keys at %q", metaPath)
	}

	keys := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		s, ok := v.(string)
		if !ok {
			continue
		}
		// Entries ending with "/" are sub-directories; include them as-is so
		// callers can decide whether to recurse.
		keys = append(keys, strings.TrimRight(s, "/"))
	}

	return keys, nil
}

func buildMetaPath(mount, path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return fmt.Sprintf("%s/metadata", mount)
	}
	return fmt.Sprintf("%s/metadata/%s", mount, path)
}
