package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/user/vaultlink/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "vaultlink.yaml")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestLoad_Valid(t *testing.T) {
	content := `
version: "1"
environments:
  - name: dev
    address: http://localhost:8200
    token: root
    prefix: secret/dev
  - name: prod
    address: https://vault.prod.example.com
    token: s.prod
    prefix: secret/prod
`
	p := writeTemp(t, content)
	cfg, err := config.Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(cfg.Environments))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/vaultlink.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestValidate_DuplicateName(t *testing.T) {
	content := `
version: "1"
environments:
  - name: dev
    address: http://localhost:8200
    token: root
  - name: dev
    address: http://localhost:8201
    token: root
`
	p := writeTemp(t, content)
	_, err := config.Load(p)
	if err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestFindEnvironment(t *testing.T) {
	cfg := &config.Config{
		Environments: []config.Environment{
			{Name: "staging", Address: "http://vault.staging", Token: "tok"},
		},
	}
	env, err := cfg.FindEnvironment("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env.Name != "staging" {
		t.Errorf("expected staging, got %s", env.Name)
	}

	_, err = cfg.FindEnvironment("missing")
	if err == nil {
		t.Fatal("expected error for missing environment")
	}
}
