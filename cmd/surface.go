package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/report"
	"github.com/mukunjin/depx/internal/surface"
	"github.com/spf13/cobra"
)

var surfaceCmd = &cobra.Command{
	Use:   "surface [path]",
	Short: "分析依赖影响面",
	Long:  `分析项目中每个依赖的影响面，包括使用频率、文件分布和关键度。`,
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

		// 检测项目类型并获取依赖列表
		m, err := analyzer.DetectManifest(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		deps, err := m.Dependencies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无法读取依赖列表 %v\n", err)
			os.Exit(1)
		}

		// 分析影响面
		results, err := surface.AnalyzeSurface(absPath, deps)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 影响面分析失败 %v\n", err)
			os.Exit(1)
		}

		// 输出报告
		report.PrintSurfaceReport(results)
	},
}

func init() {
	rootCmd.AddCommand(surfaceCmd)
}
