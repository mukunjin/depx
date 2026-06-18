package report

import (
	"fmt"
	"sort"

	"github.com/depx/depx/internal/analyzer"
	"github.com/fatih/color"
)

// PrintTerminal 在终端打印扫描结果
func PrintTerminal(result *analyzer.ScanResult) {
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	// 标题
	cyan.Println("\n  Project Summary")
	cyan.Println("--------------------------")

	// 基本信息
	fmt.Printf("  %-16s %s\n", "Path:", result.Path)
	fmt.Printf("  %-16s %s\n", "Package Manager:", result.ManifestType)

	// 依赖统计
	bold.Printf("  %-16s %d\n", "Dependencies:", result.TotalDeps)
	green.Printf("  %-16s %d\n", "Used:", result.UsedDeps)

	if result.UnusedDeps > 0 {
		red.Printf("  %-16s %d\n", "Unused:", result.UnusedDeps)
	} else {
		green.Printf("  %-16s %d\n", "Unused:", result.UnusedDeps)
	}

	// 未使用依赖列表
	if result.UnusedDeps > 0 {
		yellow.Println("\n  Unused Dependencies")
		yellow.Println("--------------------------")

		// 收集未使用的依赖并排序
		var unused []string
		for pkg, usage := range result.UsageDetails {
			if !usage.Used {
				unused = append(unused, pkg)
			}
		}
		sort.Strings(unused)

		for _, pkg := range unused {
			red.Printf("  [x] %s\n", pkg)
		}
	} else {
		green.Println("\n  [OK] All dependencies are used!")
	}

	fmt.Println()
}
