package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var sanitizeCmd = &cobra.Command{
	Use:   "sanitize [paths...]",
	Short: "Remove keys matching given prefixes or suffixes from secrets",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFile, _ := cmd.Flags().GetString("config")
		env, _ := cmd.Flags().GetString("env")
		prefixes, _ := cmd.Flags().GetStringSlice("prefix")
		suffixes, _ := cmd.Flags().GetStringSlice("suffix")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if len(prefixes) == 0 && len(suffixes) == 0 {
			return fmt.Errorf("at least one --prefix or --suffix must be specified")
		}

		cfg, err := loadConfig(cfgFile)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(cfg, env)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		results := vault.SanitizeSecrets(client, args, prefixes, suffixes, dryRun)

		hasErr := false
		for _, r := range results {
			switch r.Status {
			case "sanitized":
				fmt.Fprintf(os.Stdout, "[sanitized] %s (removed: %s)\n", r.Path, strings.Join(r.Removed, ", "))
			case "dry-run":
				fmt.Fprintf(os.Stdout, "[dry-run]   %s (would remove: %s)\n", r.Path, strings.Join(r.Removed, ", "))
			case "clean":
				fmt.Fprintf(os.Stdout, "[clean]     %s\n", r.Path)
			case "error":
				fmt.Fprintf(os.Stderr, "[error]     %s: %v\n", r.Path, r.Err)
				hasErr = true
			}
		}

		if hasErr {
			return fmt.Errorf("one or more secrets could not be sanitized")
		}
		return nil
	},
}

func init() {
	sanitizeCmd.Flags().String("env", "", "environment name (required)")
	_ = sanitizeCmd.MarkFlagRequired("env")
	sanitizeCmd.Flags().StringSlice("prefix", nil, "key prefix to remove (repeatable)")
	sanitizeCmd.Flags().StringSlice("suffix", nil, "key suffix to remove (repeatable)")
	sanitizeCmd.Flags().Bool("dry-run", false, "preview changes without writing")
	rootCmd.AddCommand(sanitizeCmd)
}
