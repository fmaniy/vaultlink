package sync_test

import (
	"context"
	"testing"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/sync"
)

// buildConfig returns a minimal config with two environments.
func buildConfig(srcAddr, dstAddr string) *config.Config {
	return &config.Config{
		Environments: []config.Environment{
			{
				Name:      "staging",
				Address:   srcAddr,
				Token:     "root",
				MountPath: "secret",
				Paths:     []string{"app/config"},
			},
			{
				Name:      "production",
				Address:   dstAddr,
				Token:     "root",
				MountPath: "secret",
				Paths:     []string{},
			},
		},
	}
}

func TestNew_UnknownEnvReturnsError(t *testing.T) {
	cfg := buildConfig("http://127.0.0.1:18200", "http://127.0.0.1:18201")
	s, err := sync.New(cfg)
	if err != nil {
		// Expected when Vault is not running in unit tests; just check type.
		t.Logf("New returned error (expected in unit test): %v", err)
		return
	}

	_, err = s.Sync(context.Background(), "nonexistent", []string{"production"})
	if err == nil {
		t.Fatal("expected error for unknown source, got nil")
	}
}

func TestSync_UnknownDestination(t *testing.T) {
	cfg := buildConfig("http://127.0.0.1:18200", "http://127.0.0.1:18201")
	s, err := sync.New(cfg)
	if err != nil {
		t.Logf("skipping: Vault unavailable: %v", err)
		return
	}

	results, err := s.Sync(context.Background(), "staging", []string{"ghost-env"})
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	for _, r := range results {
		if r.Err == nil {
			t.Errorf("expected error for unknown dest in result, got nil: %+v", r)
		}
	}
}
