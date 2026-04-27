package vault

import (
	"context"
	"fmt"
	"strings"
)

// MountInfo holds metadata about a KV secrets engine mount.
type MountInfo struct {
	Path    string
	Type    string
	Version int
}

// mounstLister is the interface used to list sys/mounts.
type mountsLister interface {
	ListMounts() (map[string]*MountOutput, error)
}

// MountOutput mirrors the fields we care about from the Vault API response.
type MountOutput struct {
	Type    string
	Options map[string]string
}

// ListKVMounts returns all KV engine mounts (v1 or v2) for the given client.
func (c *Client) ListKVMounts(ctx context.Context) ([]MountInfo, error) {
	sys := c.logical
	_ = sys // logical is used for secret ops; mounts come from the raw client

	mounts, err := c.listMountsRaw()
	if err != nil {
		return nil, fmt.Errorf("listing mounts: %w", err)
	}

	var result []MountInfo
	for path, m := range mounts {
		if m.Type != "kv" {
			continue
		}
		version := 1
		if v, ok := m.Options["version"]; ok && v == "2" {
			version = 2
		}
		result = append(result, MountInfo{
			Path:    strings.TrimSuffix(path, "/"),
			Type:    m.Type,
			Version: version,
		})
	}
	return result, nil
}

// listMountsRaw calls the Vault sys/mounts endpoint via the logical client.
func (c *Client) listMountsRaw() (map[string]*MountOutput, error) {
	secret, err := c.logical.Read("sys/mounts")
	if err != nil {
		return nil, err
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("empty response from sys/mounts")
	}

	result := make(map[string]*MountOutput)
	for key, raw := range secret.Data {
		if !strings.HasSuffix(key, "/") {
			continue
		}
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		out := &MountOutput{Options: make(map[string]string)}
		if t, ok := m["type"].(string); ok {
			out.Type = t
		}
		if opts, ok := m["options"].(map[string]interface{}); ok {
			for k, v := range opts {
				if s, ok := v.(string); ok {
					out.Options[k] = s
				}
			}
		}
		result[key] = out
	}
	return result, nil
}
