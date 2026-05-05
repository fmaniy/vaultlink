package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yourusername/vaultlink/internal/vault"
)

var (
	pinMount   string
	pinVersion string
	pinUnpin   bool
)

var pinCmd = &cobra.Command{
	Use:   "pin [paths...]",
	Short: "Pin or unpin secrets to a specific version label",
	Long: `Pin records a version label and timestamp inside a secret's data.
Unpin removes those metadata fields.

Example:
  vaultlink pin --mount secret --version v1.2.0 app/db app/cache
  vaultlink pin --mount secret --unpin app/db`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		env, err := findEnv(cfg, flagEnv)
		if err != nil {
			return err
		}

		client, err := vault.NewClient(env)
		if err != nil {
			return fmt.Errorf("vault client: %w", err)
		}

		if pinUnpin {
			for _, path := range args {
				res := vault.UnpinSecret(client, pinMount, path)
				if res.Error != nil {
					fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", path, res.Error)
				} else {
					fmt.Printf("UNPINNED  %s\n", path)
				}
			}
			return nil
		}

		if pinVersion == "" {
			return fmt.Errorf("--version is required when pinning")
		}

		results := vault.PinSecrets(client, pinMount, args, pinVersion)
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "ERROR  %s: %v\n", r.Path, r.Error)
			} else {
				fmt.Printf("PINNED  %s  @ %s\n", r.Path, r.Version)
			}
		}
		return nil
	},
}

func init() {
	pinCmd.Flags().StringVar(&pinMount, "mount", "secret", "KV mount name")
	pinCmd.Flags().StringVar(&pinVersion, "version", "", "Version label to pin")
	pinCmd.Flags().BoolVar(&pinUnpin, "unpin", false, "Remove pin metadata instead of setting it")
	pinCmd.Flags().StringVar(&flagEnv, "env", "", "Environment name from config")
	_ = pinCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(pinCmd)
}
