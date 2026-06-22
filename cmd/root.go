package cmd

import (
	"github.com/spf13/cobra"
)

// Version 版本号，由 build.ps1 通过 -ldflags 注入（基于 Git tag）
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "depx",
	Short:   "Dependency efficiency analyzer",
	Long:    `depx analyzes project dependencies to find unused packages and assess dependency surface area.`,
	Version: Version,
}

// Execute 执行根命令，返回错误
func Execute() error {
	return rootCmd.Execute()
}
