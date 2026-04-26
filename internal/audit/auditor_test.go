package audit

import (
	"testing"

	"github.com/vaultlink/internal/config"
)

func buildConfig() *config.Config {
	return &config.Config{
		Environments: []config.Environment{
			{Name: "staging", Address: "http://vault-staging:8200"},
			{Name: "production", Address: "http://vault-prod:8200"},
		},
	}
}

func TestDiff_MissingKey(t *testing.T) {
	a := New(buildConfig())
	src := map[string]string{"DB_PASS": "secret", "API_KEY": "abc"}
	dest := map[string]string{"DB_PASS": "secret"}

	report, err := a.Diff("staging", "production", src, dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MissingCount() != 1 {
		t.Errorf("expected 1 missing key, got %d", report.MissingCount())
	}
}

func TestDiff_MismatchValue(t *testing.T) {
	a := New(buildConfig())
	src := map[string]string{"DB_PASS": "secret123"}
	dest := map[string]string{"DB_PASS": "different"}

	report, err := a.Diff("staging", "production", src, dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MismatchCount() != 1 {
		t.Errorf("expected 1 mismatch, got %d", report.MismatchCount())
	}
}

func TestDiff_AllOk(t *testing.T) {
	a := New(buildConfig())
	src := map[string]string{"DB_PASS": "secret", "TOKEN": "tok"}
	dest := map[string]string{"DB_PASS": "secret", "TOKEN": "tok"}

	report, err := a.Diff("staging", "production", src, dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.MissingCount() != 0 || report.MismatchCount() != 0 {
		t.Errorf("expected no issues, got missing=%d mismatch=%d", report.MissingCount(), report.MismatchCount())
	}
}

func TestDiff_UnknownSourceEnv(t *testing.T) {
	a := New(buildConfig())
	_, err := a.Diff("unknown", "production", nil, nil)
	if err == nil {
		t.Fatal("expected error for unknown source env, got nil")
	}
}

func TestDiff_UnknownDestEnv(t *testing.T) {
	a := New(buildConfig())
	_, err := a.Diff("staging", "unknown", nil, nil)
	if err == nil {
		t.Fatal("expected error for unknown dest env, got nil")
	}
}
