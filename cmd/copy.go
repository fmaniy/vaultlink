package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	copySource string
	copyDest   string
	copyPrefix string
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy secrets from one environment to another",
	Long:  `Reads all secrets under a given prefix in the source environment and writes them to the destination environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		srcEnv, err := cfg.FindEnvironment(copySource)
		if err != nil {
			return fmt.Errorf("source environment: %w", err)
		}
		dstEnv, err := cfg.FindEnvironment(copyDest)
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

		count, err := vault.CopySecrets(srcClient, dstClient,
			srcEnv.Mount, dstEnv.Mount,
			copyPrefix, copyPrefix)
		if err != nil {
			return fmt.Errorf("copy secrets: %w", err)
		}

		log.Printf("copied %d secret(s) from %q to %q (prefix: %q)\n",
			count, copySource, copyDest, copyPrefix)
		return nil
	},
}

func init() {
	copyCmd.Flags().StringVarP(&copySource, "source", "s", "", "Source environment name (required)")
	copyCmd.Flags().StringVarP(&copyDest, "dest", "d", "", "Destination environment name (required)")
	copyCmd.Flags().StringVarP(&copyPrefix, "prefix", "p", "", "Secret path prefix to copy (required)")
	_ = copyCmd.MarkFlagRequired("source")
	_ = copyCmd.MarkFlagRequired("dest")
	_ = copyCmd.MarkFlagRequired("prefix")
	rootCmd.AddCommand(copyCmd)
}
