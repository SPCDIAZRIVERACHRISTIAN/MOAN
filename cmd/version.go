package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print build and version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("version=%s commit=%s date=%s\n", version, commit, date)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
