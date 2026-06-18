package report

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/efficiency"
	"github.com/mukunjin/depx/internal/surface"
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

// PrintEfficiencyReport 打印效率分析报告
func PrintEfficiencyReport(results map[string]*efficiency.EfficiencyResult) {
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	cyan.Println("\n  Dependency Efficiency Analysis")
	cyan.Println("==================================")

	// 按效率排序
	var sorted []*efficiency.EfficiencyResult
	for _, r := range results {
		sorted = append(sorted, r)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Efficiency < sorted[j].Efficiency
	})

	for _, r := range sorted {
		if r.Efficiency == 0 {
			continue
		}

		var colorFunc func(format string, a ...interface{}) (int, error)
		if r.Efficiency < 20 {
			colorFunc = red.Printf
		} else if r.Efficiency < 50 {
			colorFunc = yellow.Printf
		} else {
			colorFunc = fmt.Printf
		}

		colorFunc("  %s\n", r.Package)
		fmt.Printf("    Used Exports: %d\n", len(r.UsedExports))
		fmt.Printf("    Efficiency: %.1f%%\n", r.Efficiency)

		if len(r.UsedExports) > 0 && len(r.UsedExports) <= 10 {
			fmt.Printf("    Exports: %v\n", r.UsedExports)
		}
		fmt.Println()
	}
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

// PrintMonorepoReport 打印 monorepo 分析报告
func PrintMonorepoReport(result *analyzer.MonorepoResult) {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	cyan.Println("\n  Monorepo Analysis")
	cyan.Println("===================")

	fmt.Printf("  Root Path: %s\n", result.Path)
	fmt.Printf("  Workspaces: %d\n", len(result.Workspaces))
	fmt.Println()

	// 总体统计
	cyan.Println("  Overall Statistics")
	cyan.Println("  ------------------")
	fmt.Printf("    Total Dependencies: %d\n", result.TotalDeps)
	green.Printf("    Used: %d\n", result.UsedDeps)
	if result.UnusedDeps > 0 {
		yellow.Printf("    Unused: %d\n", result.UnusedDeps)
	} else {
		green.Printf("    Unused: %d\n", result.UnusedDeps)
	}
	fmt.Println()

	// 各工作区详情
	cyan.Println("  Workspace Details")
	cyan.Println("  -----------------")
	for _, ws := range result.WorkspaceResults {
		fmt.Printf("\n  [%s]\n", ws.Path)
		fmt.Printf("    Package Manager: %s\n", ws.ManifestType)
		fmt.Printf("    Dependencies: %d\n", ws.TotalDeps)
		green.Printf("    Used: %d\n", ws.UsedDeps)
		if ws.UnusedDeps > 0 {
			yellow.Printf("    Unused: %d\n", ws.UnusedDeps)

			// 列出未使用的依赖
			var unused []string
			for pkg, usage := range ws.UsageDetails {
				if !usage.Used {
					unused = append(unused, pkg)
				}
			}
			sort.Strings(unused)
			if len(unused) > 0 {
				fmt.Printf("    Unused Packages: %v\n", unused)
			}
		} else {
			green.Printf("    Unused: %d\n", ws.UnusedDeps)
		}
	}
	fmt.Println()
}
