package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"vaultlink/internal/config"
	"vaultlink/internal/vault"
)

func init() {
	var (
		envName string
		mount   string
		prefix  string
	)

	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Capture a point-in-time snapshot of secrets in an environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgPath, _ := cmd.Flags().GetString("config")
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			env, err := cfg.GetEnv(envName)
			if err != nil {
				return fmt.Errorf("unknown environment %q", envName)
			}

			token, err := vault.ResolveToken(env)
			if err != nil {
				return fmt.Errorf("resolve token: %w", err)
			}

			client, err := vault.NewClient(env.Address, token)
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			if mount == "" {
				mount = "secret"
			}

			_, results, err := vault.SnapshotSecrets(client, mount, prefix)
			if err != nil {
				return fmt.Errorf("snapshot: %w", err)
			}

			vault.PrintSnapshotReport(results)

			for _, r := range results {
				if !r.OK {
					os.Exit(1)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&envName, "env", "e", "", "Environment name (required)")
	cmd.Flags().StringVarP(&mount, "mount", "m", "secret", "KV mount path")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "Secret path prefix")
	_ = cmd.MarkFlagRequired("env")

	rootCmd.AddCommand(cmd)
}
