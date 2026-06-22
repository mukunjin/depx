package report

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/manifest"
	"github.com/mukunjin/depx/internal/surface"
)

// captureOutput 捕获标准输出（同时重定向 color 输出到 pipe）
func captureOutput(f func()) string {
	// 禁用颜色，避免 Windows 控制台 API 绕过 pipe
	color.NoColor = true
	defer func() { color.NoColor = false }()

	oldStdout := os.Stdout
	oldColorOutput := color.Output
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	f()

	w.Close()
	os.Stdout = oldStdout
	color.Output = oldColorOutput

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintTerminal(t *testing.T) {
	tests := []struct {
		name         string
		result       *analyzer.ScanResult
		showIndirect bool
		showTypePkgs bool
	}{
		{
			name: "all dependencies used",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    3,
				UsedDeps:     3,
				UnusedDeps:   0,
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash": {Package: "lodash", Used: true},
					"react":  {Package: "react", Used: true},
					"axios":  {Package: "axios", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
		},
		{
			name: "some dependencies unused",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    5,
				UsedDeps:     3,
				UnusedDeps:   2,
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash":     {Package: "lodash", Used: true},
					"react":      {Package: "react", Used: true},
					"axios":      {Package: "axios", Used: true},
					"moment":     {Package: "moment", Used: false},
					"typescript": {Package: "typescript", Used: false},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
		},
		{
			name: "all dependencies unused",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "go",
				TotalDeps:    2,
				UsedDeps:     0,
				UnusedDeps:   2,
				UsageDetails: map[string]*manifest.UsageResult{
					"github.com/pkg/errors":      {Package: "github.com/pkg/errors", Used: false},
					"github.com/sirupsen/logrus": {Package: "github.com/sirupsen/logrus", Used: false},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
		},
		{
			name: "with indirect dependencies hidden",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    2,
				UsedDeps:     2,
				UnusedDeps:   0,
				IndirectDeps: []string{"dep-a", "dep-b", "dep-c"},
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash": {Package: "lodash", Used: true},
					"react":  {Package: "react", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
		},
		{
			name: "with indirect dependencies shown",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    2,
				UsedDeps:     2,
				UnusedDeps:   0,
				IndirectDeps: []string{"dep-a", "dep-b", "dep-c"},
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash": {Package: "lodash", Used: true},
					"react":  {Package: "react", Used: true},
				},
			},
			showIndirect: true,
			showTypePkgs: false,
		},
		{
			name: "with type packages hidden",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    3,
				UsedDeps:     3,
				UnusedDeps:   0,
				TypePackages: 2,
				TypePkgNames: []string{"@types/node", "@types/react"},
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash":       {Package: "lodash", Used: true},
					"@types/node":  {Package: "@types/node", Used: true},
					"@types/react": {Package: "@types/react", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
		},
		{
			name: "with type packages shown",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    3,
				UsedDeps:     3,
				UnusedDeps:   0,
				TypePackages: 2,
				TypePkgNames: []string{"@types/node", "@types/react"},
				UsageDetails: map[string]*manifest.UsageResult{
					"lodash":       {Package: "lodash", Used: true},
					"@types/node":  {Package: "@types/node", Used: true},
					"@types/react": {Package: "@types/react", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintTerminal(tt.result, tt.showIndirect, false, tt.showTypePkgs)
			})

			// 验证输出包含关键信息
			if !strings.Contains(output, "Project Summary") {
				t.Errorf("output should contain 'Project Summary', got: %s", output)
			}
			if !strings.Contains(output, tt.result.Path) {
				t.Errorf("output should contain path %q", tt.result.Path)
			}
			if !strings.Contains(output, tt.result.ManifestType) {
				t.Errorf("output should contain manifest type %q", tt.result.ManifestType)
			}

			// 验证未使用依赖列表
			if tt.result.UnusedDeps > 0 {
				if !strings.Contains(output, "Unused Dependencies") {
					t.Errorf("output should contain 'Unused Dependencies' when there are unused deps, got: %s", output)
				}
			} else {
				if !strings.Contains(output, "All dependencies are used") {
					t.Errorf("output should contain 'All dependencies are used' when no unused deps, got: %s", output)
				}
			}

			// 验证间接依赖显示逻辑
			if len(tt.result.IndirectDeps) > 0 {
				// 始终显示数量（%-16s 格式会有填充空格）
				if !strings.Contains(output, "Indirect:") {
					t.Errorf("output should contain 'Indirect:', got: %s", output)
				}

				if tt.showIndirect {
					// 当 showIndirect 为 true 时，应该显示摘要
					if !strings.Contains(output, "Indirect Dependencies") {
						t.Errorf("output should contain 'Indirect Dependencies' section when showIndirect is true, got: %s", output)
					}
					if !strings.Contains(output, "Total:") {
						t.Errorf("output should contain 'Total:' when showIndirect is true, got: %s", output)
					}
				} else {
					// 当 showIndirect 为 false 时，不应该显示详情列表
					if strings.Contains(output, "[i]") {
						t.Errorf("output should NOT contain '[i]' entries when showIndirect is false, got: %s", output)
					}
				}
			}
		})
	}
}

func TestPrintTerminal_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		result       *analyzer.ScanResult
		showIndirect bool
		showTypePkgs bool
		check        func(string) error
	}{
		{
			name: "zero dependencies",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    0,
				UsedDeps:     0,
				UnusedDeps:   0,
				UsageDetails: map[string]*manifest.UsageResult{},
			},
			showIndirect: false,
			showTypePkgs: false,
			check: func(output string) error {
				if !strings.Contains(output, "No runtime dependencies to check") {
					return fmt.Errorf("expected 'No runtime dependencies' message")
				}
				return nil
			},
		},
		{
			name: "all used with type packages",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    2,
				UsedDeps:     2,
				UnusedDeps:   0,
				TypePackages: 1,
				TypePkgNames: []string{"@types/node"},
				UsageDetails: map[string]*manifest.UsageResult{
					"react":       {Package: "react", Used: true},
					"lodash":      {Package: "lodash", Used: true},
					"@types/node": {Package: "@types/node", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
			check: func(output string) error {
				if !strings.Contains(output, "Type Packages:") {
					return fmt.Errorf("expected 'Type Packages:' in output")
				}
				if !strings.Contains(output, "use --types to show") {
					return fmt.Errorf("expected hint to use --types")
				}
				return nil
			},
		},
		{
			name: "all used with type packages shown",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    2,
				UsedDeps:     2,
				UnusedDeps:   0,
				TypePackages: 1,
				TypePkgNames: []string{"@types/node"},
				UsageDetails: map[string]*manifest.UsageResult{
					"react":       {Package: "react", Used: true},
					"lodash":      {Package: "lodash", Used: true},
					"@types/node": {Package: "@types/node", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: true,
			check: func(output string) error {
				if !strings.Contains(output, "@types/node") {
					return fmt.Errorf("expected @types/node in output")
				}
				if !strings.Contains(output, "[T]") {
					return fmt.Errorf("expected [T] marker for type packages")
				}
				return nil
			},
		},
		{
			name: "indirect deps hidden",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    1,
				UsedDeps:     1,
				UnusedDeps:   0,
				IndirectDeps: []string{"indirect-pkg"},
				UsageDetails: map[string]*manifest.UsageResult{
					"react": {Package: "react", Used: true},
				},
			},
			showIndirect: false,
			showTypePkgs: false,
			check: func(output string) error {
				if !strings.Contains(output, "Indirect:") {
					return fmt.Errorf("expected 'Indirect:' in output")
				}
				if !strings.Contains(output, "use --indirect to show") {
					return fmt.Errorf("expected hint to use --indirect")
				}
				if strings.Contains(output, "indirect-pkg") {
					return fmt.Errorf("should not show indirect-pkg details when hidden")
				}
				return nil
			},
		},
		{
			name: "indirect deps shown",
			result: &analyzer.ScanResult{
				Path:         "/test/project",
				ManifestType: "npm",
				TotalDeps:    1,
				UsedDeps:     1,
				UnusedDeps:   0,
				IndirectDeps: []string{"indirect-pkg"},
				UsageDetails: map[string]*manifest.UsageResult{
					"react": {Package: "react", Used: true},
				},
			},
			showIndirect: true,
			showTypePkgs: false,
			check: func(output string) error {
				if !strings.Contains(output, "Indirect Dependencies") {
					return fmt.Errorf("expected 'Indirect Dependencies' section in output")
				}
				if !strings.Contains(output, "Total: 1") {
					return fmt.Errorf("expected 'Total: 1' in output")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintTerminal(tt.result, tt.showIndirect, false, tt.showTypePkgs)
			})

			if err := tt.check(output); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestPrintSurfaceReport_EmptyResults(t *testing.T) {
	output := captureOutput(func() {
		PrintSurfaceReport(map[string]*surface.SurfaceResult{}, 0, nil)
	})

	if !strings.Contains(output, "Surface Area") {
		t.Error("expected 'Surface Area' in output")
	}
	if !strings.Contains(output, "Packages:     0") {
		t.Error("expected 'Packages: 0' in output")
	}
}

func TestPrintSurfaceReport_AllCriticalities(t *testing.T) {
	results := map[string]*surface.SurfaceResult{
		"high-pkg": {
			Package:     "high-pkg",
			Files:       []string{"file1.js"},
			Modules:     []string{"mod1"},
			RefCount:    100,
			Criticality: "High",
		},
		"medium-pkg": {
			Package:     "medium-pkg",
			Files:       []string{"file2.js"},
			Modules:     []string{"mod2"},
			RefCount:    50,
			Criticality: "Medium",
		},
		"low-pkg": {
			Package:     "low-pkg",
			Files:       []string{"file3.js"},
			Modules:     []string{"mod3"},
			RefCount:    10,
			Criticality: "Low",
		},
	}

	output := captureOutput(func() {
		PrintSurfaceReport(results, 0, nil)
	})

	// 验证 Summary 部分
	if !strings.Contains(output, "High:") {
		t.Error("expected 'High:' in summary")
	}
	if !strings.Contains(output, "Medium:") {
		t.Error("expected 'Medium:' in summary")
	}
	if !strings.Contains(output, "Low:") {
		t.Error("expected 'Low:' in summary")
	}

	// 验证 Most Critical
	if !strings.Contains(output, "Most Critical") {
		t.Error("expected 'Most Critical' section")
	}
	if !strings.Contains(output, "high-pkg") {
		t.Error("expected high-pkg to be most critical")
	}

	// 验证 Runtime Surface
	if !strings.Contains(output, "Runtime Surface") {
		t.Error("expected 'Runtime Surface' section")
	}
}

func TestPrintSurfaceReport_WithIndirectSummary(t *testing.T) {
	results := map[string]*surface.SurfaceResult{
		"react": {
			Package:     "react",
			Files:       []string{"app.js"},
			Modules:     []string{"react"},
			RefCount:    10,
			Criticality: "High",
		},
	}

	parentCounts := map[string]int{
		"clsx":      5,
		"scheduler": 3,
		"tslib":     1, // 只被 1 个包引用，不应该显示
	}

	output := captureOutput(func() {
		PrintSurfaceReport(results, 100, parentCounts)
	})

	// 验证 Indirect Packages 部分
	if !strings.Contains(output, "Indirect Packages") {
		t.Error("expected 'Indirect Packages' section")
	}
	if !strings.Contains(output, "Total: 100") {
		t.Error("expected 'Total: 100'")
	}

	// 验证 Top Shared Dependencies
	if !strings.Contains(output, "Top Shared Dependencies") {
		t.Error("expected 'Top Shared Dependencies' section")
	}
	if !strings.Contains(output, "clsx") {
		t.Error("expected clsx in shared deps")
	}
	if !strings.Contains(output, "Required By: 5") {
		t.Error("expected 'Required By: 5' for clsx")
	}

	// tslib 只被 1 个包引用，不应该显示
	if strings.Contains(output, "tslib") {
		t.Error("tslib should not be shown (only 1 parent)")
	}
}

func TestPrintSurfaceReport_ZeroRefCountPackages(t *testing.T) {
	results := map[string]*surface.SurfaceResult{
		"unused-pkg": {
			Package:     "unused-pkg",
			Files:       []string{},
			Modules:     []string{},
			RefCount:    0,
			Criticality: "Low",
		},
		"used-pkg": {
			Package:     "used-pkg",
			Files:       []string{"file.js"},
			Modules:     []string{"mod"},
			RefCount:    5,
			Criticality: "High",
		},
	}

	output := captureOutput(func() {
		PrintSurfaceReport(results, 0, nil)
	})

	// 只应该显示 used-pkg（RefCount > 0）
	if !strings.Contains(output, "used-pkg") {
		t.Error("expected used-pkg in output")
	}
	if strings.Contains(output, "unused-pkg") {
		t.Error("unused-pkg (RefCount=0) should not be shown")
	}
	if !strings.Contains(output, "Packages:     1") {
		t.Error("expected 'Packages: 1' (only counting RefCount > 0)")
	}
}

func TestPrintTerminal_WithUnusedDependencies(t *testing.T) {
	result := &analyzer.ScanResult{
		Path:         "/test/project",
		ManifestType: "npm",
		TotalDeps:    3,
		UsedDeps:     1,
		UnusedDeps:   2,
		RuntimeDeps:  []string{"react"},
		UsageDetails: map[string]*manifest.UsageResult{
			"react":   {Package: "react", Used: true},
			"lodash":  {Package: "lodash", Used: false},
			"moment":  {Package: "moment", Used: false},
		},
	}

	output := captureOutput(func() {
		PrintTerminal(result, false, false, false)
	})

	// 验证 Unused Dependencies 部分
	if !strings.Contains(output, "Unused Dependencies") {
		t.Error("expected 'Unused Dependencies' section")
	}
	if !strings.Contains(output, "lodash") {
		t.Error("expected lodash in unused list")
	}
	if !strings.Contains(output, "moment") {
		t.Error("expected moment in unused list")
	}
	if !strings.Contains(output, "[x]") {
		t.Error("expected [x] marker for unused deps")
	}

	// 验证 Runtime Dependencies 部分
	if !strings.Contains(output, "Runtime Dependencies") {
		t.Error("expected 'Runtime Dependencies' section")
	}
	if !strings.Contains(output, "react") {
		t.Error("expected react in used list")
	}
	if !strings.Contains(output, "[✓]") {
		t.Error("expected [✓] marker for used deps")
	}
}

func TestPrintSurfaceReport(t *testing.T) {
	tests := []struct {
		name    string
		results map[string]*surface.SurfaceResult
	}{
		{
			name: "multiple dependencies with different criticality",
			results: map[string]*surface.SurfaceResult{
				"lodash": {
					Package:     "lodash",
					Files:       []string{"src/a.js", "src/b.js", "src/c.js"},
					Modules:     []string{"debounce", "throttle"},
					RefCount:    15,
					Criticality: "High",
				},
				"axios": {
					Package:     "axios",
					Files:       []string{"src/api.js"},
					Modules:     []string{"default"},
					RefCount:    3,
					Criticality: "Low",
				},
				"react": {
					Package:     "react",
					Files:       []string{"src/App.js", "src/index.js"},
					Modules:     []string{"useState", "useEffect"},
					RefCount:    8,
					Criticality: "Medium",
				},
			},
		},
		{
			name:    "empty results",
			results: map[string]*surface.SurfaceResult{},
		},
		{
			name: "single dependency",
			results: map[string]*surface.SurfaceResult{
				"moment": {
					Package:     "moment",
					Files:       []string{"src/utils/date.js"},
					Modules:     []string{"format", "parse"},
					RefCount:    5,
					Criticality: "Medium",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintSurfaceReport(tt.results, 0, nil)
			})

			// 验证输出包含标题
			if !strings.Contains(output, "Surface Area") {
				t.Errorf("output should contain 'Surface Area', got: %s", output)
			}

			// 验证每个依赖的信息
			for pkg, result := range tt.results {
				if !strings.Contains(output, pkg) {
					t.Errorf("output should contain package name %q", pkg)
				}
				if !strings.Contains(output, result.Criticality) {
					t.Errorf("output should contain criticality %q", result.Criticality)
				}
			}
		})
	}
}

func TestPrintSurfaceReportIndirectSummary(t *testing.T) {
	output := captureOutput(func() {
		PrintSurfaceReport(map[string]*surface.SurfaceResult{
			"react": {
				Package:     "react",
				Files:       []string{"src/App.js"},
				Modules:     []string{"react"},
				RefCount:    5,
				Criticality: "Medium",
			},
		}, 531, map[string]int{
			"clsx":      4,
			"scheduler": 3,
			"react-is":  3,
			"tslib":     3,
			"left-pad":  1,
		})
	})

	for _, expected := range []string{
		"Indirect Packages",
		"Total: 531",
		"Top Shared Dependencies",
		"clsx",
		"Required By: 4 direct packages",
		"scheduler",
		"Required By: 3 direct packages",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output should contain %q, got: %s", expected, output)
		}
	}

	if strings.Contains(output, "left-pad") {
		t.Errorf("output should not list single-parent indirect deps, got: %s", output)
	}
}
