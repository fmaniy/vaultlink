package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// envsCmd lists all configured environments.
var envsCmd = &cobra.Command{
	Use:   "envs",
	Short: "List all configured Vault environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		if loadedConfig == nil {
			return fmt.Errorf("config not loaded")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-40s %s\n", "NAME", "ADDRESS", "PREFIX")
		fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-40s %s\n",
			"---------------",
			"----------------------------------------",
			"-------------------------------")

		for _, env := range loadedConfig.Environments {
			prefix := env.Prefix
			if prefix == "" {
				prefix = "(none)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-15s %-40s %s\n",
				env.Name, env.Address, prefix)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envsCmd)
}
