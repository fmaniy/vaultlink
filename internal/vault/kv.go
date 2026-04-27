package vault

import (
	"context"
	"fmt"
	"strings"
)

// SecretData represents a map of key-value pairs stored at a Vault path.
type SecretData map[string]string

// ReadSecret reads a KV v2 secret at the given mount and path.
func (c *Client) ReadSecret(ctx context.Context, mount, path string) (SecretData, error) {
	fullPath := fmt.Sprintf("%s/data/%s", mount, strings.TrimPrefix(path, "/"))

	secret, err := c.logical.ReadWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading secret %q: %w", fullPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret not found at path %q", fullPath)
	}

	rawData, ok := secret.Data["data"]
	if !ok {
		return nil, fmt.Errorf("unexpected KV response: missing 'data' key at %q", fullPath)
	}

	raw, ok := rawData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected KV data format at %q", fullPath)
	}

	result := make(SecretData, len(raw))
	for k, v := range raw {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("non-string value for key %q at path %q", k, fullPath)
		}
		result[k] = str
	}
	return result, nil
}

// WriteSecret writes key-value pairs to a KV v2 secret at the given mount and path.
func (c *Client) WriteSecret(ctx context.Context, mount, path string, data SecretData) error {
	fullPath := fmt.Sprintf("%s/data/%s", mount, strings.TrimPrefix(path, "/"))

	payload := make(map[string]interface{}, len(data))
	for k, v := range data {
		payload[k] = v
	}

	_, err := c.logical.WriteWithContext(ctx, fullPath, map[string]interface{}{"data": payload})
	if err != nil {
		return fmt.Errorf("writing secret %q: %w", fullPath, err)
	}
	return nil
}

// ListSecrets returns the keys available under the given mount and path prefix.
func (c *Client) ListSecrets(ctx context.Context, mount, path string) ([]string, error) {
	fullPath := fmt.Sprintf("%s/metadata/%s", mount, strings.TrimPrefix(path, "/"))

	secret, err := c.logical.ListWithContext(ctx, fullPath)
	if err != nil {
		return nil, fmt.Errorf("listing secrets at %q: %w", fullPath, err)
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
		return nil, fmt.Errorf("unexpected keys format at %q", fullPath)
	}

	keys := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys, nil
}
