package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, build date, and git commit of diagtool.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("diagtool %s\n", Version)
		fmt.Printf("  Build date: %s\n", BuildDate)
		fmt.Printf("  Git commit: %s\n", GitCommit)
	},
}
