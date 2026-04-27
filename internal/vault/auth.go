package vault

import (
	"errors"
	"fmt"
	"os"
)

// AuthMethod represents a supported Vault authentication method.
type AuthMethod string

const (
	AuthToken      AuthMethod = "token"
	AuthAppRole    AuthMethod = "approle"
	AuthKubernetes AuthMethod = "kubernetes"
)

// AuthConfig holds the credentials for authenticating with Vault.
type AuthConfig struct {
	Method   AuthMethod
	Token    string
	RoleID   string
	SecretID string
	Role     string // kubernetes role
	JWTPath  string // path to the service account JWT
}

// ResolveToken returns a Vault token based on the AuthConfig.
// For token auth it returns the token directly; for approle and kubernetes
// it performs a login against the Vault API and returns the client token.
func ResolveToken(c *Client, cfg AuthConfig) (string, error) {
	switch cfg.Method {
	case AuthToken:
		if cfg.Token == "" {
			cfg.Token = os.Getenv("VAULT_TOKEN")
		}
		if cfg.Token == "" {
			return "", errors.New("auth: token is empty and VAULT_TOKEN is not set")
		}
		return cfg.Token, nil

	case AuthAppRole:
		if cfg.RoleID == "" || cfg.SecretID == "" {
			return "", errors.New("auth: approle requires role_id and secret_id")
		}
		data := map[string]interface{}{
			"role_id":   cfg.RoleID,
			"secret_id": cfg.SecretID,
		}
		secret, err := c.logical.Write("auth/approle/login", data)
		if err != nil {
			return "", fmt.Errorf("auth: approle login failed: %w", err)
		}
		if secret == nil || secret.Auth == nil {
			return "", errors.New("auth: approle login returned no auth info")
		}
		return secret.Auth.ClientToken, nil

	case AuthKubernetes:
		if cfg.Role == "" {
			return "", errors.New("auth: kubernetes requires a role")
		}
		jwtPath := cfg.JWTPath
		if jwtPath == "" {
			jwtPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		}
		jwt, err := os.ReadFile(jwtPath)
		if err != nil {
			return "", fmt.Errorf("auth: reading kubernetes JWT: %w", err)
		}
		data := map[string]interface{}{
			"role": cfg.Role,
			"jwt":  string(jwt),
		}
		secret, err := c.logical.Write("auth/kubernetes/login", data)
		if err != nil {
			return "", fmt.Errorf("auth: kubernetes login failed: %w", err)
		}
		if secret == nil || secret.Auth == nil {
			return "", errors.New("auth: kubernetes login returned no auth info")
		}
		return secret.Auth.ClientToken, nil

	default:
		return "", fmt.Errorf("auth: unsupported method %q", cfg.Method)
	}
}
