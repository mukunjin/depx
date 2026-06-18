package cmd

import (
	"fmt"
	"os"

	"github.com/depx/depx/internal/analyzer"
	"github.com/depx/depx/internal/report"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "扫描项目，检测未使用的依赖",
	Long:  `扫描指定目录下的项目，分析依赖使用情况，找出未使用的依赖。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		result, err := analyzer.Scan(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		report.PrintTerminal(result)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
