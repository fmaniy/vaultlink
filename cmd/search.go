package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	searchEnv   string
	searchMount string
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search secrets by key or value across a Vault environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		env, err := cfg.FindEnvironment(searchEnv)
		if err != nil {
			return err
		}

		token, err := vault.ResolveToken(env)
		if err != nil {
			return fmt.Errorf("resolving token: %w", err)
		}

		client, err := vault.NewClient(env.Address, token)
		if err != nil {
			return fmt.Errorf("creating vault client: %w", err)
		}

		mount := searchMount
		if mount == "" {
			mount = env.Mount
		}

		results, err := vault.SearchSecrets(client, mount, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "search error: %v\n", err)
			os.Exit(1)
		}

		vault.PrintSearchReport(query, results)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchEnv, "env", "e", "", "Environment name to search (required)")
	searchCmd.Flags().StringVarP(&searchMount, "mount", "m", "", "KV mount to search (overrides config)")
	_ = searchCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(searchCmd)
}
