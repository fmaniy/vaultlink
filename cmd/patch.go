package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	patchKey       string
	patchValue     string
	patchCreateKey bool
	patchPaths     []string
)

var patchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Update a single key within existing secrets",
	Long: `patch updates the value of a specific key across one or more secret paths.
By default the key must already exist; use --create to add new keys.`,
	Example: `  vaultlink patch --path secret/data/app --key db_pass --value s3cr3t
  vaultlink patch --path secret/data/app --path secret/data/worker --key timeout --value 30 --create`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if patchKey == "" {
			return fmt.Errorf("--key is required")
		}
		if patchValue == "" {
			return fmt.Errorf("--value is required")
		}
		if len(patchPaths) == 0 {
			return fmt.Errorf("at least one --path is required")
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		env, err := cfg.FindEnvironment(envName)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(env)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		results := vault.PatchSecrets(client, patchPaths, patchKey, patchValue, patchCreateKey)
		vault.PrintPatchReport(results)

		for _, r := range results {
			if r.Err != nil {
				os.Exit(1)
			}
		}
		return nil
	},
}

func init() {
	patchCmd.Flags().StringVar(&envName, "env", "", "environment name (required)")
	patchCmd.Flags().StringVar(&patchKey, "key", "", "key to patch within the secret")
	patchCmd.Flags().StringVar(&patchValue, "value", "", "new value for the key")
	patchCmd.Flags().StringArrayVar(&patchPaths, "path", []string{}, "secret path(s) to patch (repeatable)")
	patchCmd.Flags().BoolVar(&patchCreateKey, "create", false, "create the key if it does not exist")
	_ = patchCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(patchCmd)
}
