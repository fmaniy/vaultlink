package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/config"
	"github.com/yourusername/vaultlink/internal/vault"
)

var compareCmd = &cobra.Command{
	Use:   "compare <src-env> <dst-env> <path>",
	Short: "Compare secrets at a path between two environments",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgFile, _ := cmd.Flags().GetString("config")
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		srcName, dstName, path := args[0], args[1], args[2]

		srcEnv, err := cfg.FindEnv(srcName)
		if err != nil {
			return err
		}
		dstEnv, err := cfg.FindEnv(dstName)
		if err != nil {
			return err
		}

		srcClient, err := vault.NewClient(srcEnv)
		if err != nil {
			return fmt.Errorf("src client: %w", err)
		}
		dstClient, err := vault.NewClient(dstEnv)
		if err != nil {
			return fmt.Errorf("dst client: %w", err)
		}

		results, err := vault.CompareSecrets(srcClient, dstClient, srcEnv.Mount, dstEnv.Mount, path, path)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "PATH\tSTATUS\tDETAILS\n")
		fmt.Fprintf(w, "----\t------\t-------\n")
		matches, mismatches, missing := 0, 0, 0
		for _, r := range results {
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, r.Status, r.Details)
			switch r.Status {
			case "match":
				matches++
			case "mismatch":
				mismatches++
			default:
				missing++
			}
		}
		w.Flush()
		fmt.Printf("\nSummary: %d match, %d mismatch, %d missing\n", matches, mismatches, missing)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)
	compareCmd.Flags().String("config", "configs/vaultlink.yaml", "path to config file")
}
