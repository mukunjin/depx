package cmd

import (
	"fmt"
	"os"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/config"
	"github.com/mukunjin/depx/internal/report"
	"github.com/spf13/cobra"
)

var configPath string

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "扫描项目，检测未使用的依赖",
	Long:  `扫描指定目录下的项目，分析依赖使用情况，找出未使用的依赖。`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		return runScan(path, configPath)
	},
}

func runScan(path, configPath string) error {
	// 验证路径是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	} else {
		cfg, err = config.FindAndLoad(path)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
	}

	result, err := analyzer.ScanWithConfig(path, cfg)
	if err != nil {
		return err
	}

	report.PrintTerminal(result)
	return nil
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径 (.depx.yml)")
	rootCmd.AddCommand(scanCmd)
}
