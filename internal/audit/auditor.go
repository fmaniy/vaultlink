package audit

import (
	"fmt"
	"time"

	"github.com/vaultlink/internal/config"
)

// DiffEntry represents a single secret difference between two environments.
type DiffEntry struct {
	Key      string
	SourceEnv string
	DestEnv  string
	Status   string // "missing", "mismatch", "ok"
}

// Report holds the full audit result between two environments.
type Report struct {
	GeneratedAt time.Time
	Source      string
	Dest        string
	Entries     []DiffEntry
}

// MissingCount returns the number of keys missing in the destination.
func (r *Report) MissingCount() int {
	count := 0
	for _, e := range r.Entries {
		if e.Status == "missing" {
			count++
		}
	}
	return count
}

// MismatchCount returns the number of keys with differing values.
func (r *Report) MismatchCount() int {
	count := 0
	for _, e := range r.Entries {
		if e.Status == "mismatch" {
			count++
		}
	}
	return count
}

// Auditor compares secrets between two environments.
type Auditor struct {
	cfg *config.Config
}

// New creates a new Auditor instance.
func New(cfg *config.Config) *Auditor {
	return &Auditor{cfg: cfg}
}

// Diff compares two maps of secrets and returns a Report.
func (a *Auditor) Diff(srcName, destName string, srcSecrets, destSecrets map[string]string) (*Report, error) {
	if _, err := a.cfg.FindEnvironment(srcName); err != nil {
		return nil, fmt.Errorf("source environment %q not found: %w", srcName, err)
	}
	if _, err := a.cfg.FindEnvironment(destName); err != nil {
		return nil, fmt.Errorf("destination environment %q not found: %w", destName, err)
	}

	report := &Report{
		GeneratedAt: time.Now(),
		Source:      srcName,
		Dest:        destName,
	}

	for key, srcVal := range srcSecrets {
		destVal, exists := destSecrets[key]
		switch {
		case !exists:
			report.Entries = append(report.Entries, DiffEntry{Key: key, SourceEnv: srcName, DestEnv: destName, Status: "missing"})
		case srcVal != destVal:
			report.Entries = append(report.Entries, DiffEntry{Key: key, SourceEnv: srcName, DestEnv: destName, Status: "mismatch"})
		default:
			report.Entries = append(report.Entries, DiffEntry{Key: key, SourceEnv: srcName, DestEnv: destName, Status: "ok"})
		}
	}

	return report, nil
}
