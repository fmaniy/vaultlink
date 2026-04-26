package sync

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

// Result holds the outcome of a single secret sync operation.
type Result struct {
	Path    string
	Source  string
	Dest    string
	Action  string // "created", "updated", "skipped"
	Err     error
}

// Syncer copies secrets from a source environment to one or more destinations.
type Syncer struct {
	clients map[string]*vault.Client
	cfg     *config.Config
}

// New creates a Syncer, initialising a Vault client for every environment.
func New(cfg *config.Config) (*Syncer, error) {
	clients := make(map[string]*vault.Client, len(cfg.Environments))
	for _, env := range cfg.Environments {
		c, err := vault.NewClient(env)
		if err != nil {
			return nil, fmt.Errorf("syncer: init client for %q: %w", env.Name, err)
		}
		clients[env.Name] = c
	}
	return &Syncer{clients: clients, cfg: cfg}, nil
}

// Sync copies all secrets listed under the source environment to each
// destination environment and returns one Result per path per destination.
func (s *Syncer) Sync(ctx context.Context, sourceName string, destNames []string) ([]Result, error) {
	src, ok := s.clients[sourceName]
	if !ok {
		return nil, fmt.Errorf("syncer: unknown source environment %q", sourceName)
	}

	srcEnv, err := s.cfg.FindEnvironment(sourceName)
	if err != nil {
		return nil, err
	}

	var results []Result
	for _, path := range srcEnv.Paths {
		secrets, err := src.ReadSecret(ctx, path)
		if err != nil {
			log.Printf("syncer: read %q from %q: %v", path, sourceName, err)
			results = append(results, Result{Path: path, Source: sourceName, Err: err})
			continue
		}

		for _, destName := range destNames {
			dst, ok := s.clients[destName]
			if !ok {
				results = append(results, Result{Path: path, Source: sourceName, Dest: destName,
					Err: fmt.Errorf("unknown destination environment %q", destName)})
				continue
			}
			action, err := dst.WriteSecret(ctx, path, secrets)
			results = append(results, Result{
				Path: path, Source: sourceName, Dest: destName, Action: action, Err: err,
			})
		}
	}
	return results, nil
}
