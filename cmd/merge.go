package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	mergeOverwrite bool
	mergeDryRun    bool
	mergeSrcMount  string
	mergeDstMount  string
)

var mergeCmd = &cobra.Command{
	Use:   "merge <path> [path...]",
	Short: "Merge secrets from one KV mount into another",
	Long: `Merge reads each secret from the source mount and merges its keys
into the corresponding path in the destination mount.

Existing keys in the destination are preserved unless --overwrite is set.
Use --dry-run to preview changes without writing.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		srcEnv, err := cfg.FindEnv(mergeSrcMount)
		if err != nil {
			return fmt.Errorf("source environment not found: %w", err)
		}
		dstEnv, err := cfg.FindEnv(mergeDstMount)
		if err != nil {
			return fmt.Errorf("destination environment not found: %w", err)
		}

		srcClient, err := vault.NewClient(srcEnv)
		if err != nil {
			return fmt.Errorf("connect to source: %w", err)
		}
		dstClient, err := vault.NewClient(dstEnv)
		if err != nil {
			return fmt.Errorf("connect to destination: %w", err)
		}

		_ = srcClient
		_ = dstClient

		opts := vault.MergeOptions{
			Overwrite: mergeOverwrite,
			DryRun:    mergeDryRun,
		}

		results := vault.MergeSecrets(srcClient, srcEnv.Mount, args, dstEnv.Mount, opts)

		hasErr := false
		for _, r := range results {
			switch r.Status {
			case "merged":
				if len(r.Conflicts) > 0 {
					fmt.Fprintf(os.Stdout, "  merged  %s (conflicts: %v)\n", r.Path, r.Conflicts)
				} else {
					fmt.Fprintf(os.Stdout, "  merged  %s\n", r.Path)
				}
			case "skipped":
				fmt.Fprintf(os.Stdout, "  dry-run %s\n", r.Path)
			case "error":
				fmt.Fprintf(os.Stderr, "  error   %s: %v\n", r.Path, r.Error)
				hasErr = true
			}
		}

		if hasErr {
			return fmt.Errorf("one or more secrets failed to merge")
		}
		return nil
	},
}

func init() {
	mergeCmd.Flags().StringVar(&mergeSrcMount, "src", "", "Source environment name (required)")
	mergeCmd.Flags().StringVar(&mergeDstMount, "dst", "", "Destination environment name (required)")
	mergeCmd.Flags().BoolVar(&mergeOverwrite, "overwrite", false, "Overwrite conflicting keys in destination")
	mergeCmd.Flags().BoolVar(&mergeDryRun, "dry-run", false, "Preview changes without writing")
	_ = mergeCmd.MarkFlagRequired("src")
	_ = mergeCmd.MarkFlagRequired("dst")
	rootCmd.AddCommand(mergeCmd)
}
