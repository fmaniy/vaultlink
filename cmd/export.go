package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	exportEnv    string
	exportMount  string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export [paths...]",
	Short: "Export secrets from a Vault environment to a JSON file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		env, err := cfg.FindEnvironment(exportEnv)
		if err != nil {
			return fmt.Errorf("environment %q not found", exportEnv)
		}

		token, err := vault.ResolveToken(env)
		if err != nil {
			return fmt.Errorf("resolving token: %w", err)
		}

		client, err := vault.NewClient(env.Address, token)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		var results []vault.ExportResult
		if exportOutput != "" {
			results, err = vault.ExportSecretsToFile(client, exportMount, args, exportOutput)
		} else {
			results, err = vault.ExportSecrets(client, exportMount, args, os.Stdout)
		}
		if err != nil {
			return fmt.Errorf("exporting secrets: %w", err)
		}

		successCount, failCount := 0, 0
		for _, r := range results {
			if r.Success {
				successCount++
			} else {
				failCount++
				fmt.Fprintf(os.Stderr, "WARN: failed to export %q: %s\n", r.Path, r.Error)
			}
		}
		fmt.Fprintf(os.Stderr, "Export complete: %d succeeded, %d failed\n", successCount, failCount)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportEnv, "env", "e", "", "Source environment name (required)")
	exportCmd.Flags().StringVarP(&exportMount, "mount", "m", "secret", "KV mount path")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (defaults to stdout)")
	_ = exportCmd.MarkFlagRequired("env")
}
