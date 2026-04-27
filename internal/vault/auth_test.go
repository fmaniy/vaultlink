package vault

import (
	"errors"
	"os"
	"testing"

	vaultapi "github.com/hashicorp/vault/api"
)

func TestResolveToken_TokenFromConfig(t *testing.T) {
	c := &Client{logical: &fakeLogical{}}
	cfg := AuthConfig{Method: AuthToken, Token: "s.abc123"}
	tok, err := ResolveToken(c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "s.abc123" {
		t.Errorf("expected s.abc123, got %s", tok)
	}
}

func TestResolveToken_TokenFromEnv(t *testing.T) {
	t.Setenv("VAULT_TOKEN", "s.fromenv")
	c := &Client{logical: &fakeLogical{}}
	cfg := AuthConfig{Method: AuthToken}
	tok, err := ResolveToken(c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "s.fromenv" {
		t.Errorf("expected s.fromenv, got %s", tok)
	}
}

func TestResolveToken_TokenEmpty(t *testing.T) {
	os.Unsetenv("VAULT_TOKEN")
	c := &Client{logical: &fakeLogical{}}
	cfg := AuthConfig{Method: AuthToken}
	_, err := ResolveToken(c, cfg)
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestResolveToken_AppRole_MissingFields(t *testing.T) {
	c := &Client{logical: &fakeLogical{}}
	cfg := AuthConfig{Method: AuthAppRole, RoleID: "role"}
	_, err := ResolveToken(c, cfg)
	if err == nil {
		t.Fatal("expected error for missing secret_id")
	}
}

func TestResolveToken_AppRole_Success(t *testing.T) {
	fl := &fakeLogical{
		writeSecret: &vaultapi.Secret{
			Auth: &vaultapi.SecretAuth{ClientToken: "s.approle"},
		},
	}
	c := &Client{logical: fl}
	cfg := AuthConfig{Method: AuthAppRole, RoleID: "rid", SecretID: "sid"}
	tok, err := ResolveToken(c, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "s.approle" {
		t.Errorf("expected s.approle, got %s", tok)
	}
}

func TestResolveToken_AppRole_WriteError(t *testing.T) {
	fl := &fakeLogical{writeErr: errors.New("vault down")}
	c := &Client{logical: fl}
	cfg := AuthConfig{Method: AuthAppRole, RoleID: "rid", SecretID: "sid"}
	_, err := ResolveToken(c, cfg)
	if err == nil {
		t.Fatal("expected error from write")
	}
}

func TestResolveToken_UnsupportedMethod(t *testing.T) {
	c := &Client{logical: &fakeLogical{}}
	cfg := AuthConfig{Method: "ldap"}
	_, err := ResolveToken(c, cfg)
	if err == nil {
		t.Fatal("expected error for unsupported method")
	}
}
