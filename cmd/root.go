package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version 版本号，通过 ldflags 注入
var Version = "v0.2.0"

var rootCmd = &cobra.Command{
	Use:   "depx",
	Short: "Dependency Efficiency Analyzer",
	Long:  `depx 是一个依赖效率分析工具，帮助开发者发现未使用的依赖。`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
