package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/sync"
)

var (
	syncSource string
	syncDests  []string
	dryRun     bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets from a source environment to one or more destinations",
	Example: `  vaultlink sync --source staging --dest production
  vaultlink sync --source staging --dest production --dest qa`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if syncSource == "" {
			return fmt.Errorf("--source is required")
		}
		if len(syncDests) == 0 {
			return fmt.Errorf("at least one --dest is required")
		}

		if dryRun {
			fmt.Printf("[dry-run] would sync %q → %s\n", syncSource, strings.Join(syncDests, ", "))
			return nil
		}

		s, err := sync.New(cfg)
		if err != nil {
			return fmt.Errorf("init syncer: %w", err)
		}

		results, err := s.Sync(cmd.Context(), syncSource, syncDests)
		if err != nil {
			return fmt.Errorf("sync: %w", err)
		}

		sync.PrintReport(os.Stdout, results)
		created, updated, skipped, failed := sync.Summary(results)
		fmt.Printf("\nSummary: %d created, %d updated, %d skipped, %d failed\n",
			created, updated, skipped, failed)
		if failed > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	syncCmd.Flags().StringVar(&syncSource, "source", "", "Source environment name")
	syncCmd.Flags().StringArrayVar(&syncDests, "dest", nil, "Destination environment name(s) (repeatable)")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print what would be synced without making changes")
	rootCmd.AddCommand(syncCmd)
}
