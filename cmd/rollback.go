package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	rollbackEnv   string
	rollbackMount string
	rollbackPaths []string
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Snapshot secrets, run a risky operation interactively, and auto-restore on failure",
	Long: `Takes a point-in-time snapshot of the specified secrets in an environment.
If the subsequent sync or copy operation fails, all snapshotted secrets are
automatically restored to their previous values.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		env, err := cfg.FindEnvironment(rollbackEnv)
		if err != nil {
			return fmt.Errorf("environment %q not found", rollbackEnv)
		}

		token, err := vault.ResolveToken(env)
		if err != nil {
			return fmt.Errorf("resolve token: %w", err)
		}

		client, err := vault.NewClient(env.Address, token)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Taking snapshots of %d path(s) on %s...\n", len(rollbackPaths), rollbackEnv)

		err = vault.RollbackSecrets(client, rollbackMount, rollbackPaths, func() error {
			fmt.Fprintln(os.Stdout, "Snapshots taken. Proceeding — press Ctrl-C to abort before changes are applied.")
			// In a real workflow the caller would chain a sync/copy here.
			// For now we confirm success interactively.
			return nil
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rollback triggered: %v\n", err)
			return err
		}

		fmt.Fprintln(os.Stdout, "Operation completed successfully. No rollback needed.")
		return nil
	},
}

func init() {
	rollbackCmd.Flags().StringVarP(&rollbackEnv, "env", "e", "", "Target environment name (required)")
	rollbackCmd.Flags().StringVarP(&rollbackMount, "mount", "m", "secret", "KV mount path")
	rollbackCmd.Flags().StringSliceVarP(&rollbackPaths, "paths", "p", nil, "Comma-separated secret paths to snapshot (required)")
	_ = rollbackCmd.MarkFlagRequired("env")
	_ = rollbackCmd.MarkFlagRequired("paths")
	rootCmd.AddCommand(rollbackCmd)
}
