package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/analyzer"
)

// TestIntegration_NpmProject 测试 npm 项目扫描
func TestIntegration_NpmProject(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "npm-project")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "npm" {
		t.Errorf("Expected manifest type 'npm', got '%s'", result.ManifestType)
	}

	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}

	// 验证结果结构完整
	if result.Path == "" {
		t.Error("Expected non-empty path")
	}

	if result.UsageDetails == nil {
		t.Error("Expected non-nil UsageDetails")
	}

	// 验证统计正确
	if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
		t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
			result.UsedDeps, result.UnusedDeps, result.TotalDeps)
	}
}

// TestIntegration_GoProject 测试 Go 项目扫描
func TestIntegration_GoProject(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "go-project")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "go" {
		t.Errorf("Expected manifest type 'go', got '%s'", result.ManifestType)
	}

	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}
}

// TestIntegration_RustProject 测试 Rust 项目扫描
func TestIntegration_RustProject(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "rust-project")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "cargo" {
		t.Errorf("Expected manifest type 'cargo', got '%s'", result.ManifestType)
	}

	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}
}

// TestIntegration_PythonProject 测试 Python 项目扫描
func TestIntegration_PythonProject(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "python-project")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "pip" {
		t.Errorf("Expected manifest type 'pip', got '%s'", result.ManifestType)
	}

	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}
}

// TestIntegration_ComplexNpmProject 测试复杂 npm 项目
func TestIntegration_ComplexNpmProject(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "npm-complex")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// 复杂项目应该有更多依赖
	if result.TotalDeps < 5 {
		t.Errorf("Expected at least 5 dependencies for complex project, got %d", result.TotalDeps)
	}

	// 验证 UsageDetails 包含详细信息
	for pkg, usage := range result.UsageDetails {
		if usage.Package != pkg {
			t.Errorf("Package name mismatch: key='%s', usage.Package='%s'", pkg, usage.Package)
		}
	}
}

// TestIntegration_EdgeCases 测试边界情况
func TestIntegration_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{
			name:        "all used",
			dir:         filepath.Join("..", "testdata", "edge-all-used"),
			expectError: false,
		},
		{
			name:        "none used",
			dir:         filepath.Join("..", "testdata", "edge-none-used"),
			expectError: false,
		},
		{
			name:        "no source files",
			dir:         filepath.Join("..", "testdata", "edge-no-source"),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := os.Stat(tt.dir); os.IsNotExist(err) {
				t.Skipf("testdata not found: %s", tt.dir)
			}

			result, err := analyzer.Scan(tt.dir)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			// 验证统计正确
			if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
				t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
					result.UsedDeps, result.UnusedDeps, result.TotalDeps)
			}
		})
	}
}

// TestIntegration_RealWorldNpm 测试真实 npm 项目结构
func TestIntegration_RealWorldNpm(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "real-npm")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// 验证多文件扫描
	hasUsedIn := false
	for _, usage := range result.UsageDetails {
		if usage.Used && len(usage.UsedIn) > 0 {
			hasUsedIn = true
			break
		}
	}

	if !hasUsedIn && result.UsedDeps > 0 {
		t.Error("Expected UsedIn to be populated for used dependencies")
	}
}

// TestIntegration_GoComplex 测试复杂 Go 项目
func TestIntegration_GoComplex(t *testing.T) {
	testdataDir := filepath.Join("..", "testdata", "go-complex")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// 复杂项目应该有测试文件，但不应该影响依赖分析
	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies in complex Go project")
	}
}
