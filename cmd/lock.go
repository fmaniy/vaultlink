package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var lockCmd = &cobra.Command{
	Use:   "lock [mount] [paths...]",
	Short: "Lock secrets to prevent unintended modification",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		mount := args[0]
		paths := args[1:]

		lockedBy, _ := cmd.Flags().GetString("by")
		if lockedBy == "" {
			lockedBy = os.Getenv("USER")
			if lockedBy == "" {
				lockedBy = "unknown"
			}
		}

		unlock, _ := cmd.Flags().GetBool("unlock")

		cfg, err := loadConfig(cmd)
		if err != nil {
			return err
		}
		env, err := cfg.FindEnvironment(envName)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(env)
		if err != nil {
			return fmt.Errorf("failed to create vault client: %w", err)
		}

		if unlock {
			for _, p := range paths {
				res := vault.UnlockSecret(client, mount, p)
				if res.Success {
					fmt.Printf("✔ unlocked: %s\n", p)
				} else {
					fmt.Fprintf(os.Stderr, "✘ unlock failed [%s]: %v\n", p, res.Error)
				}
			}
			return nil
		}

		results := vault.LockSecrets(client, mount, paths, lockedBy)
		for _, r := range results {
			if r.Success {
				fmt.Printf("✔ locked: %s (by %s)\n", r.Path, lockedBy)
			} else {
				fmt.Fprintf(os.Stderr, "✘ lock failed [%s]: %v\n", r.Path, r.Error)
			}
		}
		return nil
	},
}

func init() {
	lockCmd.Flags().String("by", "", "Identity of the person/system locking the secret (defaults to $USER)")
	lockCmd.Flags().Bool("unlock", false, "Remove the lock marker from the specified secrets")
	lockCmd.Flags().StringVar(&envName, "env", "", "Environment name from config")
	_ = lockCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(lockCmd)
}
