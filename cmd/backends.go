package cmd

import (
	"fmt"

	"github.com/cwbudde/matplotlib-go/backends"
	_ "github.com/cwbudde/matplotlib-go/backends/all" // register every built-in backend
	"github.com/cwbudde/matplotlib-go/render"
	"github.com/spf13/cobra"
)

var backendsCmd = &cobra.Command{
	Use:   "backends",
	Short: "Inspect registered rendering backends",
	Long:  "List installed rendering backends and their declared capabilities.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(backends.CapabilityMatrix())
	},
}

var backendsCompareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Report runtime native/fallback/unsupported capability status per backend",
	Long: `Instantiates each available backend with a small headless configuration
and prints a unsupported/fallback/native status matrix. Use this to verify
that backend-declared capabilities match runtime renderer interface support.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := backends.Config{
			Width:      128,
			Height:     128,
			Background: render.Color{R: 1, G: 1, B: 1, A: 1},
			DPI:        72,
		}
		fmt.Println(backends.BackendComparisonReport(config))
	},
}

func init() {
	backendsCmd.AddCommand(backendsCompareCmd)
	rootCmd.AddCommand(backendsCmd)
}
