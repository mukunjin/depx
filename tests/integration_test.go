package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/analyzer"
)

// scanTestdata 扫描 testdata 目录并返回结果
func scanTestdata(t *testing.T, dir string) *analyzer.ScanResult {
	t.Helper()
	testdataDir := filepath.Join("..", "testdata", dir)
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	result, err := analyzer.Scan(testdataDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	return result
}

// assertScanResult 验证扫描结果的基本完整性
func assertScanResult(t *testing.T, result *analyzer.ScanResult, expectedType string) {
	t.Helper()
	if result.ManifestType != expectedType {
		t.Errorf("Expected manifest type '%s', got '%s'", expectedType, result.ManifestType)
	}
	if result.TotalDeps == 0 {
		t.Error("Expected some dependencies, got 0")
	}
	if result.Path == "" {
		t.Error("Expected non-empty path")
	}
	if result.UsageDetails == nil {
		t.Error("Expected non-nil UsageDetails")
	}
	if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
		t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
			result.UsedDeps, result.UnusedDeps, result.TotalDeps)
	}
}

// TestIntegration_NpmProject 测试 npm 项目扫描
func TestIntegration_NpmProject(t *testing.T) {
	result := scanTestdata(t, "npm-project")
	assertScanResult(t, result, "npm")
}

// TestIntegration_GoProject 测试 Go 项目扫描
func TestIntegration_GoProject(t *testing.T) {
	result := scanTestdata(t, "go-project")
	assertScanResult(t, result, "go")
}

// TestIntegration_RustProject 测试 Rust 项目扫描
func TestIntegration_RustProject(t *testing.T) {
	result := scanTestdata(t, "rust-project")
	assertScanResult(t, result, "cargo")
}

// TestIntegration_PythonProject 测试 Python 项目扫描
func TestIntegration_PythonProject(t *testing.T) {
	result := scanTestdata(t, "python-project")
	assertScanResult(t, result, "pip")
}

// TestIntegration_ComplexNpmProject 测试复杂 npm 项目
func TestIntegration_ComplexNpmProject(t *testing.T) {
	result := scanTestdata(t, "npm-complex")
	assertScanResult(t, result, "npm")

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
		name string
		dir  string
	}{
		{name: "all used", dir: "edge-all-used"},
		{name: "none used", dir: "edge-none-used"},
		{name: "no source files", dir: "edge-no-source"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scanTestdata(t, tt.dir)
			if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
				t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
					result.UsedDeps, result.UnusedDeps, result.TotalDeps)
			}
		})
	}
}

// TestIntegration_RealWorldNpm 测试真实 npm 项目结构
func TestIntegration_RealWorldNpm(t *testing.T) {
	result := scanTestdata(t, "real-npm")
	assertScanResult(t, result, "npm")

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
	result := scanTestdata(t, "go-complex")
	assertScanResult(t, result, "go")
}

// TestIntegration_RustComplex 测试复杂 Rust 项目
func TestIntegration_RustComplex(t *testing.T) {
	result := scanTestdata(t, "rust-complex")
	assertScanResult(t, result, "cargo")

	// 验证多文件扫描
	if result.TotalDeps < 6 {
		t.Errorf("Expected at least 6 dependencies for complex Rust project, got %d", result.TotalDeps)
	}

	// 验证 serde 和 tokio 被使用（在 main.rs 和 handlers.rs 中）
	if usage, ok := result.UsageDetails["serde"]; ok {
		if !usage.Used {
			t.Error("serde should be marked as used")
		}
		if len(usage.UsedIn) < 1 {
			t.Error("serde should be used in at least 1 file")
		}
	} else {
		t.Error("serde not found in results")
	}

	if usage, ok := result.UsageDetails["tokio"]; ok {
		if !usage.Used {
			t.Error("tokio should be marked as used")
		}
	} else {
		t.Error("tokio not found in results")
	}

	// 验证 unused-crate 未被使用
	if usage, ok := result.UsageDetails["unused-crate"]; ok {
		if usage.Used {
			t.Error("unused-crate should not be marked as used")
		}
	} else {
		t.Error("unused-crate not found in results")
	}
}

// TestIntegration_PythonComplex 测试复杂 Python 项目
func TestIntegration_PythonComplex(t *testing.T) {
	result := scanTestdata(t, "python-complex")
	assertScanResult(t, result, "pip")

	// 验证多文件扫描
	if result.TotalDeps < 6 {
		t.Errorf("Expected at least 6 dependencies for complex Python project, got %d", result.TotalDeps)
	}

	// 验证 requests 和 flask 被使用
	if usage, ok := result.UsageDetails["requests"]; ok {
		if !usage.Used {
			t.Error("requests should be marked as used")
		}
		if len(usage.UsedIn) < 1 {
			t.Error("requests should be used in at least 1 file")
		}
	} else {
		t.Error("requests not found in results")
	}

	if usage, ok := result.UsageDetails["flask"]; ok {
		if !usage.Used {
			t.Error("flask should be marked as used")
		}
	} else {
		t.Error("flask not found in results")
	}

	// 验证 numpy 被使用（在 models.py 中）
	if usage, ok := result.UsageDetails["numpy"]; ok {
		if !usage.Used {
			t.Error("numpy should be marked as used")
		}
	} else {
		t.Error("numpy not found in results")
	}

	// 验证 sqlalchemy 和 redis 被使用（在 database.py 中）
	if usage, ok := result.UsageDetails["sqlalchemy"]; ok {
		if !usage.Used {
			t.Error("sqlalchemy should be marked as used")
		}
	} else {
		t.Error("sqlalchemy not found in results")
	}

	if usage, ok := result.UsageDetails["redis"]; ok {
		if !usage.Used {
			t.Error("redis should be marked as used")
		}
	} else {
		t.Error("redis not found in results")
	}

	// 验证 unused-package 未被使用
	if usage, ok := result.UsageDetails["unused-package"]; ok {
		if usage.Used {
			t.Error("unused-package should not be marked as used")
		}
	} else {
		t.Error("unused-package not found in results")
	}
}
