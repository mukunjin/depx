package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/config"
	"github.com/mukunjin/depx/internal/report"
	"github.com/mukunjin/depx/internal/surface"
	"github.com/spf13/cobra"
)

var surfaceConfigPath string

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

		return runSurface(path, surfaceConfigPath)
	},
}

func runSurface(path, configPath string) error {
	// 验证路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path %s: %w", path, err)
	}

	// 加载配置
	var cfg *config.Config
	if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	} else {
		cfg, err = config.FindAndLoad(absPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
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

	// 过滤被忽略的依赖
	var filteredDeps []string
	for _, dep := range deps {
		if !cfg.IsIgnored(dep) {
			filteredDeps = append(filteredDeps, dep)
		}
	}

	// 构建 surface options
	opts := &surface.Options{
		ExcludeDirs:     cfg.ExcludeDirs,
		ExcludeFiles:    cfg.ExcludeFiles,
		ReadNodeModules: cfg.ReadNodeModules,
	}

	// 分析影响面
	results, err := surface.AnalyzeSurface(absPath, filteredDeps, opts)
	if err != nil {
		return err
	}

	// 输出报告
	report.PrintSurfaceReport(results)
	return nil
}

func init() {
	surfaceCmd.Flags().StringVarP(&surfaceConfigPath, "config", "c", "", "配置文件路径 (.depx.yml)")
	rootCmd.AddCommand(surfaceCmd)
}
