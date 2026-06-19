package cmd

import (
	"github.com/spf13/cobra"
)

// Version 版本号，由 build.ps1 通过 -ldflags 注入（基于 Git tag）
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "depx",
	Short:   "Dependency Efficiency Analyzer",
	Long:    `depx 是一个依赖效率分析工具，帮助开发者发现未使用的依赖。`,
	Version: Version,
}

// Execute 执行根命令，返回错误
func Execute() error {
	return rootCmd.Execute()
}
