package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	rotatePaths []string
	rotateKeys  []string
	rotateMount string
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate secrets by re-writing them with a new generated value",
	Long: `Rotate one or more secrets at the given paths.
All specified keys (or all keys if none given) are overwritten with a
timestamp-based placeholder, forcing a new secret version in Vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(rotatePaths) == 0 {
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

		results, summary := vault.RotateSecrets(client, rotateMount, rotatePaths, rotateKeys)

		fmt.Fprintf(os.Stdout, "%-40s %s\n", "PATH", "STATUS")
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 55))
		for _, r := range results {
			status := "OK"
			if !r.Success {
				status = fmt.Sprintf("FAILED (%v)", r.Error)
			}
			fmt.Fprintf(os.Stdout, "%-40s %s\n", r.Path, status)
		}
		fmt.Fprintln(os.Stdout, strings.Repeat("-", 55))
		fmt.Fprintf(os.Stdout, "Total: %d  Rotated: %d  Failed: %d\n",
			summary.Total, summary.Rotated, summary.Failed)

		if summary.Failed > 0 {
			return fmt.Errorf("%d secret(s) failed to rotate", summary.Failed)
		}
		return nil
	},
}

func init() {
	rotateCmd.Flags().StringVarP(&envName, "env", "e", "", "environment name (required)")
	rotateCmd.Flags().StringVar(&rotateMount, "mount", "secret", "KV mount path")
	rotateCmd.Flags().StringArrayVar(&rotatePaths, "path", nil, "secret path(s) to rotate (repeatable)")
	rotateCmd.Flags().StringArrayVar(&rotateKeys, "key", nil, "specific key(s) to rotate; omit for all keys")
	_ = rotateCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(rotateCmd)
}
