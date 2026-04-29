package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	promoteOverwrite bool
	promoteMount     string
)

var promoteCmd = &cobra.Command{
	Use:   "promote <source-env> <dest-env> [path...]",
	Short: "Promote secrets from one environment to another",
	Args:  cobra.MinimumNArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcName, dstName, paths := args[0], args[1], args[2:]

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		srcEnv, err := cfg.FindEnvironment(srcName)
		if err != nil {
			return fmt.Errorf("source environment: %w", err)
		}
		dstEnv, err := cfg.FindEnvironment(dstName)
		if err != nil {
			return fmt.Errorf("destination environment: %w", err)
		}

		srcClient, err := vault.NewClient(srcEnv)
		if err != nil {
			return fmt.Errorf("source vault client: %w", err)
		}
		dstClient, err := vault.NewClient(dstEnv)
		if err != nil {
			return fmt.Errorf("destination vault client: %w", err)
		}

		results := vault.PromoteSecrets(srcClient, dstClient, paths, promoteOverwrite)
		vault.PrintPromoteReport(results, srcName, dstName)

		for _, r := range results {
			if !r.Success {
				os.Exit(1)
			}
		}
		return nil
	},
}

func init() {
	promoteCmd.Flags().BoolVar(&promoteOverwrite, "overwrite", false, "Overwrite existing secrets in the destination")
	promoteCmd.Flags().StringVar(&promoteMount, "mount", "secret", "KV mount path")
	rootCmd.AddCommand(promoteCmd)
}
