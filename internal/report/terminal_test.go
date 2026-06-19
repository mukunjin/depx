package report

import (
	"bytes"
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
		name   string
		result *analyzer.ScanResult
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(func() {
				PrintTerminal(tt.result)
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
		})
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
				PrintSurfaceReport(tt.results)
			})

			// 验证输出包含标题
			if !strings.Contains(output, "Dependency Surface Area Analysis") {
				t.Errorf("output should contain 'Dependency Surface Area Analysis', got: %s", output)
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
