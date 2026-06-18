package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/report"
	"github.com/spf13/cobra"
)

var monorepoCmd = &cobra.Command{
	Use:   "monorepo [path]",
	Short: "分析 monorepo 项目",
	Long:  `分析 monorepo 项目中所有工作区的依赖使用情况。`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// 获取绝对路径
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法解析路径 %s\n", err)
			os.Exit(1)
		}

		// 扫描 monorepo
		result, err := analyzer.ScanMonorepo(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: monorepo 分析失败 %v\n", err)
			os.Exit(1)
		}

		// 输出报告
		report.PrintMonorepoReport(result)
	},
}

func init() {
	rootCmd.AddCommand(monorepoCmd)
}
