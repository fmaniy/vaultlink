package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var revertCmd = &cobra.Command{
	Use:   "revert [mount] [path...]",
	Short: "Revert one or more secrets to a previous version",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mount := args[0]
		paths := args[1:]

		version, _ := cmd.Flags().GetInt("version")
		env, _ := cmd.Flags().GetString("env")

		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(cfg, env)
		if err != nil {
			return fmt.Errorf("create vault client: %w", err)
		}

		results := vault.RevertSecrets(client, mount, paths, version)

		ok, failed := 0, 0
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "  FAIL  %s: %v\n", r.Path, r.Error)
				failed++
			} else {
				fmt.Printf("  OK    %s (reverted to v%d)\n", r.Path, version)
				ok++
			}
		}

		fmt.Printf("\nSummary: %d reverted, %d failed\n", ok, failed)
		if failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	revertCmd.Flags().Int("version", 1, "Secret version to revert to")
	revertCmd.Flags().String("env", "", "Environment name from config")
	revertCmd.Flags().String("config", "", "Path to config file")
	rootCmd.AddCommand(revertCmd)
}
