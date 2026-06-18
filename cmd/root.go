package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "depx",
	Short: "Dependency Efficiency Analyzer",
	Long:  `depx 是一个依赖效率分析工具，帮助开发者发现未使用的依赖。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
