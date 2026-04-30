package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	watchMount    string
	watchInterval int
)

var watchCmd = &cobra.Command{
	Use:   "watch <path>",
	Short: "Watch a secret path and print changes in real time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("creating vault client: %w", err)
		}

		path := args[0]
		duration := time.Duration(watchInterval) * time.Second

		fmt.Fprintf(os.Stdout, "Watching %s/%s every %s — press Ctrl+C to stop\n", watchMount, path, duration)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		ch := vault.WatchSecret(ctx, client, watchMount, path, duration)
		for result := range ch {
			ts := result.Timestamp.Format("15:04:05")
			if result.Err != nil {
				fmt.Fprintf(os.Stderr, "[%s] ERROR: %v\n", ts, result.Err)
				continue
			}
			if result.Changed {
				fmt.Fprintf(os.Stdout, "[%s] CHANGED %s\n", ts, result.Path)
				for k, v := range result.Data {
					fmt.Fprintf(os.Stdout, "  %s = %v\n", k, v)
				}
			} else {
				fmt.Fprintf(os.Stdout, "[%s] no change\n", ts)
			}
		}
		return nil
	},
}

func init() {
	watchCmd.Flags().StringVarP(&envName, "env", "e", "", "environment name (required)")
	watchCmd.Flags().StringVarP(&watchMount, "mount", "m", "secret", "KV mount to use")
	watchCmd.Flags().IntVarP(&watchInterval, "interval", "i", 10, "polling interval in seconds")
	_ = watchCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(watchCmd)
}
