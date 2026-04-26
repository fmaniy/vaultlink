// Package vault provides a thin wrapper around the HashiCorp Vault API client,
// scoped to the operations vaultlink needs: reading, writing, listing, and
// deleting KV v2 secrets.
package vault

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the official Vault API client and exposes higher-level helpers
// that work exclusively with the KV v2 secrets engine.
type Client struct {
	raw    *vaultapi.Client
	mount  string // KV v2 mount path, e.g. "secret"
}

// NewClient creates an authenticated Vault client for the given address and
// token. mount is the KV v2 engine mount path (defaults to "secret" if empty).
func NewClient(address, token, mount string) (*Client, error) {
	if address == "" {
		return nil, fmt.Errorf("vault address must not be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("vault token must not be empty")
	}
	if mount == "" {
		mount = "secret"
	}

	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	raw, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}
	raw.SetToken(token)

	return &Client{raw: raw, mount: mount}, nil
}

// GetSecret retrieves the latest version of a KV v2 secret at path.
// It returns the key/value data map or an error if the path does not exist.
func (c *Client) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	kvPath := c.kvDataPath(path)
	secret, err := c.raw.KVv2(c.mount).Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading secret %q (resolved: %q): %w", path, kvPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("secret %q not found", path)
	}

	result := make(map[string]string, len(secret.Data))
	for k, v := range secret.Data {
		str, ok := v.(string)
		if !ok {
			str = fmt.Sprintf("%v", v)
		}
		result[k] = str
	}
	return result, nil
}

// PutSecret writes (or updates) a KV v2 secret at path with the provided data.
func (c *Client) PutSecret(ctx context.Context, path string, data map[string]string) error {
	raw := make(map[string]interface{}, len(data))
	for k, v := range data {
		raw[k] = v
	}
	_, err := c.raw.KVv2(c.mount).Put(ctx, path, raw)
	if err != nil {
		return fmt.Errorf("writing secret %q: %w", path, err)
	}
	return nil
}

// ListSecrets returns the keys (paths) available under the given prefix.
// For KV v2 the prefix should not include the mount or "metadata/" segments.
func (c *Client) ListSecrets(ctx context.Context, prefix string) ([]string, error) {
	// The underlying SDK list call works on the metadata path.
	metaPath := strings.TrimRight(c.mount+"/metadata/"+strings.TrimLeft(prefix, "/"), "/")
	secret, err := c.raw.Logical().ListWithContext(ctx, metaPath)
	if err != nil {
		return nil, fmt.Errorf("listing secrets under %q: %w", prefix, err)
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
		return nil, fmt.Errorf("unexpected type for keys field: %T", raw)
	}

	keys := make([]string, 0, len(ifaces))
	for _, v := range ifaces {
		if s, ok := v.(string); ok {
			keys = append(keys, s)
		}
	}
	return keys, nil
}

// kvDataPath returns the full KV v2 data path for the given logical path.
// Useful for error messages; actual API calls use the SDK helpers.
func (c *Client) kvDataPath(path string) string {
	return c.mount + "/data/" + strings.TrimLeft(path, "/")
}
