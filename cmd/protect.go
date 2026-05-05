package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	protectMount     string
	protectPaths     []string
	protectUnprotect bool
)

var protectCmd = &cobra.Command{
	Use:   "protect",
	Short: "Mark secrets as protected to prevent accidental overwrites",
	Example: `  vaultlink protect --mount secret --path apps/prod/db --path apps/prod/api
  vaultlink protect --mount secret --path apps/prod/db --unprotect`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(protectPaths) == 0 {
			return fmt.Errorf("at least one --path is required")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		env, err := resolveEnv(cfg)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(env)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		results := vault.ProtectSecrets(client, protectMount, protectPaths, protectUnprotect)

		action := "protected"
		if protectUnprotect {
			action = "unprotected"
		}

		failed := 0
		for _, r := range results {
			switch {
			case r.Error != nil:
				fmt.Fprintf(os.Stderr, "  ERROR   %s: %v\n", r.Path, r.Error)
				failed++
			case r.Skipped:
				fmt.Printf("  SKIPPED %s (already %s)\n", r.Path, action)
			default:
				fmt.Printf("  OK      %s\n", r.Path)
			}
		}

		if failed > 0 {
			return fmt.Errorf("%d path(s) failed", failed)
		}
		return nil
	},
}

func init() {
	protectCmd.Flags().StringVar(&protectMount, "mount", "secret", "KV mount to target")
	protectCmd.Flags().StringArrayVar(&protectPaths, "path", nil, "Secret path(s) to protect (repeatable)")
	protectCmd.Flags().BoolVar(&protectUnprotect, "unprotect", false, "Remove protection flag instead of setting it")
	rootCmd.AddCommand(protectCmd)
}
