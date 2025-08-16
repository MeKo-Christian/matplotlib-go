package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// These are set via -ldflags at build time, e.g.:
//
//	-X 'matplotlib-go/cmd.Version=v0.0.1' -X 'matplotlib-go/cmd.Commit=abcdef' -X 'matplotlib-go/cmd.Date=2025-08-16T00:00:00Z'
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		v := Version
		if v == "" {
			v = "dev"
		}
		if Commit != "" && Date != "" {
			fmt.Printf("matplotlib-go %s (commit %s, %s)\n", v, Commit, Date)
			return
		}
		if Commit != "" {
			fmt.Printf("matplotlib-go %s (commit %s)\n", v, Commit)
			return
		}
		fmt.Printf("matplotlib-go %s\n", v)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
