package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	annotateEnv   string
	annotateMount string
	annotatePaths []string
	annotateKey   string
	annotateValue string
)

var annotateCmd = &cobra.Command{
	Use:   "annotate",
	Short: "Set a metadata annotation on one or more secrets",
	Example: `  vaultlink annotate --env prod --mount secret --path app/db --key owner --value team-a
  vaultlink annotate --env prod --mount secret --path app/db --path app/api --key env --value prod`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if annotateKey == "" || annotateValue == "" {
			return fmt.Errorf("--key and --value are required")
		}
		if len(annotatePaths) == 0 {
			return fmt.Errorf("at least one --path is required")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		client, err := vault.NewClient(cfg, annotateEnv)
		if err != nil {
			return fmt.Errorf("create vault client: %w", err)
		}

		fullPaths := make([]string, len(annotatePaths))
		for i, p := range annotatePaths {
			fullPaths[i] = annotateMount + "/data/" + p
		}

		results := vault.AnnotateSecrets(client, fullPaths, annotateKey, annotateValue)
		vault.PrintAnnotateReport(results)

		for _, r := range results {
			if r.Err != nil {
				os.Exit(1)
			}
		}
		return nil
	},
}

func init() {
	annotateCmd.Flags().StringVar(&annotateEnv, "env", "", "target environment (required)")
	annotateCmd.Flags().StringVar(&annotateMount, "mount", "secret", "KV mount name")
	annotateCmd.Flags().StringArrayVar(&annotatePaths, "path", nil, "secret path(s) to annotate (repeatable)")
	annotateCmd.Flags().StringVar(&annotateKey, "key", "", "annotation key (required)")
	annotateCmd.Flags().StringVar(&annotateValue, "value", "", "annotation value (required)")
	_ = annotateCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(annotateCmd)
}
