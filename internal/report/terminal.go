package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/surface"
)

// typesPrefix 是 TypeScript 类型包的统一前缀
const typesPrefix = "@types/"

// PrintTerminal 在终端打印扫描结果
func PrintTerminal(result *analyzer.ScanResult, showIndirect, showAllIndirect, showTypePkgs bool) {
	bold := color.New(color.FgWhite, color.Bold)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	gray := color.New(color.FgHiBlack)

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

	// 类型包统计（默认折叠，只显示数量）
	if result.TypePackages > 0 {
		if showTypePkgs {
			red.Printf("  %-16s %d\n", "Type Packages:", result.TypePackages)
		} else {
			gray.Printf("  %-16s %d (use --types to show)\n", "Type Packages:", result.TypePackages)
		}
	}

	// 间接依赖统计（默认折叠，只显示数量）
	if len(result.IndirectDeps) > 0 {
		if showIndirect || showAllIndirect {
			cyan.Printf("  %-16s %d\n", "Indirect:", len(result.IndirectDeps))
		} else {
			gray.Printf("  %-16s %d (use --indirect to show)\n", "Indirect:", len(result.IndirectDeps))
		}
	}

	// Runtime Dependencies（已使用的运行时依赖）
	runtimeUsed := make([]string, 0)
	for _, pkg := range result.RuntimeDeps {
		if usage, ok := result.UsageDetails[pkg]; ok && usage.Used && !strings.HasPrefix(pkg, typesPrefix) {
			runtimeUsed = append(runtimeUsed, pkg)
		}
	}
	
	if len(runtimeUsed) > 0 {
		green.Println("\n  Runtime Dependencies")
		green.Println("--------------------------")
		sort.Strings(runtimeUsed)
		for _, pkg := range runtimeUsed {
			green.Printf("  [✓] %s\n", pkg)
		}
	}

	// Tool Packages（已使用的工具包）
	toolUsed := make([]string, 0)
	for _, pkg := range result.ToolDeps {
		if usage, ok := result.UsageDetails[pkg]; ok && usage.Used && !strings.HasPrefix(pkg, typesPrefix) {
			toolUsed = append(toolUsed, pkg)
		}
	}
	
	if len(toolUsed) > 0 {
		cyan.Println("\n  Tool Packages")
		cyan.Println("--------------------------")
		sort.Strings(toolUsed)
		for _, pkg := range toolUsed {
			cyan.Printf("  [✓] %s\n", pkg)
		}
	}

	// 未使用依赖列表（排除类型包）
	if result.UnusedDeps > 0 {
		yellow.Println("\n  Unused Dependencies")
		yellow.Println("--------------------------")

		// 收集未使用的依赖并排序（排除 @types/ 开头的包）
		var unused []string
		for pkg, usage := range result.UsageDetails {
			if !usage.Used && !strings.HasPrefix(pkg, typesPrefix) {
				unused = append(unused, pkg)
			}
		}
		sort.Strings(unused)

		for _, pkg := range unused {
			red.Printf("  [x] %s\n", pkg)
		}
	} else if result.UsedDeps == 0 {
		green.Println("\n  [OK] No runtime dependencies to check!")
	} else {
		green.Println("\n  [OK] All dependencies are used!")
	}

	// 类型包详情（仅在 showTypePkgs 为 true 时显示）
	if showTypePkgs && result.TypePackages > 0 {
		cyan.Println("\n  Type Packages")
		cyan.Println("--------------------------")

		// 排序类型包
		typePkgsSorted := make([]string, len(result.TypePkgNames))
		copy(typePkgsSorted, result.TypePkgNames)
		sort.Strings(typePkgsSorted)

		for _, pkg := range typePkgsSorted {
			fmt.Printf("  [T] %s\n", pkg)
		}
	}

	// 间接依赖详情
	if len(result.IndirectDeps) > 0 && (showIndirect || showAllIndirect) {
		cyan.Println("\n  Indirect Dependencies")
		cyan.Println("--------------------------")

		if showAllIndirect {
			// 显示全部间接依赖
			indirectSorted := make([]string, len(result.IndirectDeps))
			copy(indirectSorted, result.IndirectDeps)
			sort.Strings(indirectSorted)

			for _, pkg := range indirectSorted {
				fmt.Printf("  [i] %s\n", pkg)
			}
		} else {
			// 摘要模式：显示总数 + Top Shared
			fmt.Printf("  Total: %d\n\n", len(result.IndirectDeps))

			// 显示 Top Shared（被 2 个以上直接依赖引用的间接依赖）
			if result.IndirectSharedCounts != nil {
				type sharedDep struct {
					name  string
					count int
				}
				var shared []sharedDep
				for name, count := range result.IndirectSharedCounts {
					if count >= 2 {
						shared = append(shared, sharedDep{name: name, count: count})
					}
				}

				if len(shared) > 0 {
					sort.Slice(shared, func(i, j int) bool {
						if shared[i].count != shared[j].count {
							return shared[i].count > shared[j].count
						}
						return shared[i].name < shared[j].name
					})

					cyan.Println("  Top Shared")
					cyan.Println("  --------------------------")
					
					// 最多显示 10 个
					limit := 10
					if len(shared) < limit {
						limit = len(shared)
					}
					for i := 0; i < limit; i++ {
						fmt.Printf("  %-20s (%d parents)\n", shared[i].name, shared[i].count)
					}
					
					if len(shared) > 10 {
						fmt.Printf("  ... and %d more\n", len(shared)-10)
					}
					
					fmt.Println()
				}
			}

			gray.Println("  Use --indirect-all to show all packages.")
		}
	}

	fmt.Println()
}

// PrintSurfaceReport 打印影响面分析报告
func PrintSurfaceReport(runtimeResults map[string]*surface.SurfaceResult, indirectTotal int, parentCounts map[string]int) {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.FgWhite, color.Bold)

	cyan.Println("\n  Surface Area")
	cyan.Println("===================================")

	// 计算统计数据（排除 RefCount=0 的包）
	var usedPackages []*surface.SurfaceResult
	highCount, mediumCount, lowCount := 0, 0, 0
	for _, result := range runtimeResults {
		if result.RefCount > 0 {
			usedPackages = append(usedPackages, result)
			switch result.Criticality {
			case "High":
				highCount++
			case "Medium":
				mediumCount++
			case "Low":
				lowCount++
			}
		}
	}

	// 显示 Summary
	cyan.Println("\n  Summary")
	cyan.Println("--------------------------")
	bold.Printf("  Packages:     %d\n", len(usedPackages))
	if highCount > 0 {
		red.Printf("  High:         %d\n", highCount)
	} else {
		fmt.Printf("  High:         %d\n", highCount)
	}
	if mediumCount > 0 {
		yellow.Printf("  Medium:       %d\n", mediumCount)
	} else {
		fmt.Printf("  Medium:       %d\n", mediumCount)
	}
	if lowCount > 0 {
		green.Printf("  Low:          %d\n", lowCount)
	} else {
		fmt.Printf("  Low:          %d\n", lowCount)
	}

	// 显示 Most Critical (分数最高的包)
	if len(usedPackages) > 0 {
		// 找出分数最高的包
		topPkg := usedPackages[0]
		for _, pkg := range usedPackages {
			if pkg.Score > topPkg.Score {
				topPkg = pkg
			}
		}

		cyan.Println("\n  Most Critical")
		cyan.Println("--------------------------")
		bold.Printf("  %s\n", topPkg.Package)
		fmt.Printf("  Score: %d\n", topPkg.Score)
	}

	// 显示 Runtime Surface（只显示 RefCount > 0 的包）
	cyan.Println("\n  Runtime Surface")
	cyan.Println("--------------------------")
	printSurfaceResults(usedPackages, green, yellow, red)

	// 显示 Indirect Packages 摘要（如果有）
	if indirectTotal > 0 {
		cyan.Println("\n  Indirect Packages")
		cyan.Println("--------------------------")
		fmt.Printf("\n  Total: %d\n", indirectTotal)

		type sharedDep struct {
			name  string
			count int
		}
		var shared []sharedDep
		for name, count := range parentCounts {
			if count >= 2 {
				shared = append(shared, sharedDep{name: name, count: count})
			}
		}

		if len(shared) == 0 {
			fmt.Println("\n  (No shared indirect dependencies detected)")
			fmt.Println()
			return
		}

		sort.Slice(shared, func(i, j int) bool {
			if shared[i].count != shared[j].count {
				return shared[i].count > shared[j].count
			}
			return shared[i].name < shared[j].name
		})

		cyan.Println("\n  Top Shared Dependencies")
		cyan.Println("--------------------------")
		for _, item := range shared {
			fmt.Printf("\n  %s\n", item.name)
			fmt.Printf("    Required By: %d direct packages\n", item.count)
		}
	}

	fmt.Println()
}

// printSurfaceResults 打印影响面结果
func printSurfaceResults(results []*surface.SurfaceResult, green, yellow, red *color.Color) {
	// 按 Score 降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	for _, r := range results {
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
