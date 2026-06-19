package cmd

import (
	"fmt"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		return runSurface(path)
	},
}

func runSurface(path string) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path %s: %w", path, err)
	}

	// 检测项目类型并获取依赖列表
	m, err := analyzer.DetectManifest(absPath)
	if err != nil {
		return err
	}

	deps, err := m.Dependencies()
	if err != nil {
		return fmt.Errorf("reading dependencies: %w", err)
	}

	// 分析影响面
	results, err := surface.AnalyzeSurface(absPath, deps)
	if err != nil {
		return err
	}

	// 输出报告
	report.PrintSurfaceReport(results)
	return nil
}

func init() {
	rootCmd.AddCommand(surfaceCmd)
}
