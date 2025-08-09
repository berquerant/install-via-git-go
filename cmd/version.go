package cmd

import (
	"berquerant/install-via-git-go/version"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, _ []string) {
		version.Write(cmd.OutOrStdout())
	},
}
