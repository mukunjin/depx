package cmd

import (
	"fmt"
	"os"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/report"
	"github.com/spf13/cobra"
)

var (
	configPath    string
	showIndirect  bool
	showAllIndirect bool
	showTypePkgs  bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan project for unused dependencies",
	Long: `Scan the project directory, analyze dependency usage in source files,
and report unused dependencies. Both runtime and dev dependencies are checked.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		return runScan(path, configPath, showIndirect, showAllIndirect, showTypePkgs)
	},
}

func runScan(path, configPath string, showIndirect, showAllIndirect, showTypePkgs bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	cfg, err := loadConfig(path, configPath)
	if err != nil {
		return err
	}

	result, err := analyzer.ScanWithConfig(path, cfg)
	if err != nil {
		return err
	}

	report.PrintTerminal(result, showIndirect, showAllIndirect, showTypePkgs)
	return nil
}

func init() {
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (.depx.yml)")
	scanCmd.Flags().BoolVarP(&showIndirect, "indirect", "i", false, "Show indirect dependency summary")
	scanCmd.Flags().BoolVar(&showAllIndirect, "indirect-all", false, "Show all indirect dependencies")
	scanCmd.Flags().BoolVarP(&showTypePkgs, "types", "t", false, "Show type package details")
	rootCmd.AddCommand(scanCmd)
}
