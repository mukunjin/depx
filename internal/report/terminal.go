package report

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/surface"
)

// PrintTerminal 在终端打印扫描结果
func PrintTerminal(result *analyzer.ScanResult, showIndirect bool) {
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

	// 显示间接依赖数量
	if len(result.IndirectDeps) > 0 {
		cyan.Printf("  %-16s %d\n", "Indirect:", len(result.IndirectDeps))
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

	// 间接依赖列表（仅在 showIndirect 为 true 时显示）
	if showIndirect && len(result.IndirectDeps) > 0 {
		cyan.Println("\n  Indirect Dependencies")
		cyan.Println("--------------------------")

		// 排序间接依赖
		indirectSorted := make([]string, len(result.IndirectDeps))
		copy(indirectSorted, result.IndirectDeps)
		sort.Strings(indirectSorted)

		for _, pkg := range indirectSorted {
			fmt.Printf("  [i] %s\n", pkg)
		}
	}

	fmt.Println()
}

// PrintSurfaceReport 打印影响面分析报告
func PrintSurfaceReport(results map[string]*surface.SurfaceResult) {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	cyan.Println("\n  Dependency Surface Area Analysis")
	cyan.Println("===================================")

	// 按关键度排序
	var sorted []*surface.SurfaceResult
	for _, r := range results {
		sorted = append(sorted, r)
	}
	criticalityOrder := map[string]int{"High": 3, "Medium": 2, "Low": 1}
	sort.Slice(sorted, func(i, j int) bool {
		return criticalityOrder[sorted[i].Criticality] > criticalityOrder[sorted[j].Criticality]
	})

	for _, r := range sorted {
		var critColor *color.Color
		switch r.Criticality {
		case "High":
			critColor = red
		case "Medium":
			critColor = yellow
		default:
			critColor = green
		}

		fmt.Printf("  %s\n", r.Package)
		critColor.Printf("    Criticality: %s\n", r.Criticality)
		fmt.Printf("    Files: %d\n", len(r.Files))
		fmt.Printf("    Modules: %d\n", len(r.Modules))
		fmt.Printf("    Ref Count: %d\n", r.RefCount)
		fmt.Println()
	}
}
