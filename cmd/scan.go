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
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		var cfg *config.Config
		var err error

		if configPath != "" {
			cfg, err = config.Load(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
		} else {
			cfg, err = config.FindAndLoad(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not load config: %v\n", err)
				cfg = config.DefaultConfig()
			}
		}

		result, err := analyzer.ScanWithConfig(path, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		report.PrintTerminal(result)
	},
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径 (.depx.yml)")
	rootCmd.AddCommand(scanCmd)
}
