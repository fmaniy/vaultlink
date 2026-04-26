package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/vaultlink/internal/config"
)

var (
	cfgFile string
	loadedConfig *config.Config
)

// rootCmd is the base command for the vaultlink CLI.
var rootCmd = &cobra.Command{
	Use:   "vaultlink",
	Short: "Sync and audit HashiCorp Vault secrets across environments",
	Long: `vaultlink is a CLI tool that helps you sync secrets between
HashiCorp Vault environments and audit differences across them.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}
		var err error
		loadedConfig, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		return nil
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&cfgFile,
		"config", "c",
		"vaultlink.yaml",
		"path to vaultlink config file",
	)
}
