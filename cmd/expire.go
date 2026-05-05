package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

func init() {
	var mount string
	var paths []string
	var expiredOnly bool

	cmd := &cobra.Command{
		Use:   "expire",
		Short: "Check expiry status of secrets tagged with __expires_at",
		Example: `  vaultlink expire --mount secret --path myapp/db --path myapp/token`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgFile, _ := cmd.Flags().GetString("config")
			env, _ := cmd.Flags().GetString("env")

			client, err := vault.NewClient(cfgFile, env)
			if err != nil {
				return fmt.Errorf("vault client: %w", err)
			}

			results := vault.CheckExpiry(client, mount, paths, time.Now().UTC())

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "PATH\tSTATUS\tEXPIRES AT")
			fmt.Fprintln(w, "----\t------\t----------")

			expiredCount := 0
			for _, r := range results {
				if r.Error != nil {
					fmt.Fprintf(w, "%s\tERROR\t%v\n", r.Path, r.Error)
					continue
				}
				if r.Missing {
					fmt.Fprintf(w, "%s\tMISSING\t-\n", r.Path)
					continue
				}
				status := "OK"
				expAt := "none"
				if !r.ExpiresAt.IsZero() {
					expAt = r.ExpiresAt.Format(time.RFC3339)
					if r.Expired {
						status = "EXPIRED"
						expiredCount++
					}
				}
				if expiredOnly && !r.Expired {
					continue
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", r.Path, status, expAt)
			}
			w.Flush()
			fmt.Fprintf(os.Stdout, "\n%d expired / %d checked\n", expiredCount, len(results))

			// Exit with a non-zero status code when expired secrets are found,
			// allowing callers (e.g. CI pipelines) to detect expiry automatically.
			if expiredCount > 0 {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&mount, "mount", "secret", "KV mount to check")
	cmd.Flags().StringArrayVar(&paths, "path", nil, "Secret path(s) to check (repeatable)")
	cmd.Flags().BoolVar(&expiredOnly, "expired-only", false, "Only show expired secrets")
	_ = cmd.MarkFlagRequired("path")

	rootCmd.AddCommand(cmd)
}
