package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/analyzer"
	"github.com/mukunjin/depx/internal/surface"
	"github.com/mukunjin/depx/tests/helpers"
)

// 使用 tests/helpers 中的共享函数

// TestIntegration_NpmProject 测试 npm 项目扫描
func TestIntegration_NpmProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "npm-project")
	helpers.AssertScanResult(t, result, "npm")

	// 验证具体依赖（4 个 dependencies + 2 个 devDependencies = 6）
	if result.TotalDeps != 6 {
		t.Errorf("Expected 6 total deps, got %d", result.TotalDeps)
	}

	// axios 和 lodash 应该被使用
	if usage, ok := result.UsageDetails["axios"]; ok {
		if !usage.Used {
			t.Error("axios should be used")
		}
	} else {
		t.Error("axios not found")
	}

	if usage, ok := result.UsageDetails["lodash"]; ok {
		if !usage.Used {
			t.Error("lodash should be used")
		}
	} else {
		t.Error("lodash not found")
	}

	// moment 和 chalk 未使用
	if usage, ok := result.UsageDetails["moment"]; ok {
		if usage.Used {
			t.Error("moment should not be used")
		}
	} else {
		t.Error("moment not found")
	}
}

// TestIntegration_GoProject 测试 Go 项目扫描
func TestIntegration_GoProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "go-project")
	helpers.AssertScanResult(t, result, "go")

	// 验证具体依赖（indirect 依赖被过滤，只计 2 个直接依赖）
	if result.TotalDeps != 2 {
		t.Errorf("Expected 2 total deps, got %d", result.TotalDeps)
	}

	// gin 应该被使用
	if usage, ok := result.UsageDetails["github.com/gin-gonic/gin"]; ok {
		if !usage.Used {
			t.Error("gin should be used")
		}
	} else {
		t.Error("gin not found")
	}

	// cobra 未使用
	if usage, ok := result.UsageDetails["github.com/spf13/cobra"]; ok {
		if usage.Used {
			t.Error("cobra should not be used")
		}
	} else {
		t.Error("cobra not found")
	}
}

// TestIntegration_RustProject 测试 Rust 项目扫描
func TestIntegration_RustProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "rust-project")
	helpers.AssertScanResult(t, result, "cargo")

	// 验证具体依赖
	if result.TotalDeps != 3 {
		t.Errorf("Expected 3 total deps, got %d", result.TotalDeps)
	}

	// serde, tokio, reqwest 都应该被使用
	for _, pkg := range []string{"serde", "tokio", "reqwest"} {
		if usage, ok := result.UsageDetails[pkg]; ok {
			if !usage.Used {
				t.Errorf("%s should be used", pkg)
			}
		} else {
			t.Errorf("%s not found", pkg)
		}
	}
}

// TestIntegration_PythonProject 测试 Python 项目扫描
func TestIntegration_PythonProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "python-project")
	helpers.AssertScanResult(t, result, "pip")

	// 验证具体依赖
	if result.TotalDeps != 4 {
		t.Errorf("Expected 4 total deps, got %d", result.TotalDeps)
	}

	// requests, flask, numpy, pandas 都应该被使用
	for _, pkg := range []string{"requests", "flask", "numpy", "pandas"} {
		if usage, ok := result.UsageDetails[pkg]; ok {
			if !usage.Used {
				t.Errorf("%s should be used", pkg)
			}
		} else {
			t.Errorf("%s not found", pkg)
		}
	}
}

// TestIntegration_ComplexNpmProject 测试复杂 npm 项目
func TestIntegration_ComplexNpmProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "npm-complex")
	helpers.AssertScanResult(t, result, "npm")

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
	t.Parallel()
	tests := []struct {
		name           string
		dir            string
		expectedUsed   int
		expectedUnused int
	}{
		{name: "all used", dir: "edge-all-used", expectedUsed: 2, expectedUnused: 0},
		{name: "none used", dir: "edge-none-used", expectedUsed: 0, expectedUnused: 3},
		{name: "no source files", dir: "edge-no-source", expectedUsed: 0, expectedUnused: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helpers.ScanTestdata(t, tt.dir)
			if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
				t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
					result.UsedDeps, result.UnusedDeps, result.TotalDeps)
			}
			if result.UsedDeps != tt.expectedUsed {
				t.Errorf("Expected %d used deps, got %d", tt.expectedUsed, result.UsedDeps)
			}
			if result.UnusedDeps != tt.expectedUnused {
				t.Errorf("Expected %d unused deps, got %d", tt.expectedUnused, result.UnusedDeps)
			}
		})
	}
}

// TestIntegration_RealWorldNpm 测试真实 npm 项目结构
func TestIntegration_RealWorldNpm(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "real-npm")
	helpers.AssertScanResult(t, result, "npm")

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
	t.Parallel()
	result := helpers.ScanTestdata(t, "go-complex")
	helpers.AssertScanResult(t, result, "go")
}

// TestIntegration_RustComplex 测试复杂 Rust 项目
func TestIntegration_RustComplex(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "rust-complex")
	helpers.AssertScanResult(t, result, "cargo")

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
	t.Parallel()
	result := helpers.ScanTestdata(t, "python-complex")
	helpers.AssertScanResult(t, result, "pip")

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

// TestIntegration_NonExistentPath 测试扫描不存在的项目路径
func TestIntegration_NonExistentPath(t *testing.T) {
	t.Parallel()
	nonExistentPath := filepath.Join("..", "testdata", "this-path-does-not-exist")

	_, err := analyzer.Scan(nonExistentPath)
	if err == nil {
		t.Error("Expected error when scanning non-existent path, got nil")
	}
}

// TestIntegration_DefaultConfig 测试扫描没有配置文件的项目（应使用默认配置）
func TestIntegration_DefaultConfig(t *testing.T) {
	t.Parallel()
	// npm-project 没有 .depx.yml 配置文件
	result := helpers.ScanTestdata(t, "npm-project")

	// 应该成功扫描并使用默认配置
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// 默认配置应该启用 lockfile 分析
	if result.ManifestType != "npm" {
		t.Errorf("Expected manifest type 'npm', got '%s'", result.ManifestType)
	}

	// 验证所有依赖都被正确分析
	if result.TotalDeps != 6 {
		t.Errorf("Expected 6 total deps, got %d", result.TotalDeps)
	}
}

// TestIntegration_WithConfig 测试扫描有配置文件的项目
func TestIntegration_WithConfig(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "config-project")
	helpers.AssertScanResult(t, result, "npm")

	// 配置文件忽略了 moment 和 typescript，所以应该只有 2 个依赖被分析
	// (axios 和 lodash)
	if result.TotalDeps != 2 {
		t.Errorf("Expected 2 total deps (after ignore filter), got %d", result.TotalDeps)
	}

	// 验证被忽略的依赖不在结果中
	if _, ok := result.UsageDetails["moment"]; ok {
		t.Error("moment should be ignored and not in results")
	}
	if _, ok := result.UsageDetails["typescript"]; ok {
		t.Error("typescript should be ignored and not in results")
	}

	// 验证 axios 和 lodash 在结果中
	if _, ok := result.UsageDetails["axios"]; !ok {
		t.Error("axios should be in results")
	}
	if _, ok := result.UsageDetails["lodash"]; !ok {
		t.Error("lodash should be in results")
	}
}

// TestIntegration_SurfaceCommand 测试 surface 命令的完整流程
func TestIntegration_SurfaceCommand(t *testing.T) {
	t.Parallel()
	testdataDir := filepath.Join("..", "testdata", "npm-project")
	if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
		t.Skipf("testdata not found: %s", testdataDir)
	}

	// 获取依赖列表
	m, err := analyzer.DetectManifest(testdataDir)
	if err != nil {
		t.Fatalf("DetectManifest failed: %v", err)
	}

	deps, err := m.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 验证依赖列表
	if len(deps) != 6 {
		t.Errorf("Expected 6 deps, got %d", len(deps))
	}

	// 分析影响面
	opts := &surface.Options{
		ExcludeDirs:     []string{"node_modules"},
		ExcludeFiles:    []string{},
		ReadNodeModules: false,
	}

	results, err := surface.AnalyzeSurface(testdataDir, deps, opts)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 验证结果
	if len(results) != 6 {
		t.Errorf("Expected 6 surface results, got %d", len(results))
	}

	// 验证 axios 的影响面
	if axiosResult, ok := results["axios"]; ok {
		if len(axiosResult.Files) == 0 {
			t.Error("axios should be used in at least 1 file")
		}
		if axiosResult.RefCount == 0 {
			t.Error("axios should have at least 1 reference")
		}
		if axiosResult.Criticality == "" {
			t.Error("axios should have a criticality level")
		}
	} else {
		t.Error("axios not found in surface results")
	}

	// 验证 moment 的影响面（应该没有被使用）
	if momentResult, ok := results["moment"]; ok {
		if len(momentResult.Files) != 0 {
			t.Error("moment should not be used in any files")
		}
		if momentResult.RefCount != 0 {
			t.Error("moment should have 0 references")
		}
	} else {
		t.Error("moment not found in surface results")
	}
}

// TestIntegration_MixedProject 测试混合项目（多个 manifest 文件）
func TestIntegration_MixedProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "complex-mixed")

	// 应该检测到 npm 项目（优先级最高）
	if result.ManifestType != "npm" {
		t.Errorf("Expected manifest type 'npm' (highest priority), got '%s'", result.ManifestType)
	}

	// 应该只分析 npm 依赖
	if result.TotalDeps != 2 {
		t.Errorf("Expected 2 npm deps, got %d", result.TotalDeps)
	}

	// 验证 express 和 lodash 被分析
	if _, ok := result.UsageDetails["express"]; !ok {
		t.Error("express should be in results")
	}
	if _, ok := result.UsageDetails["lodash"]; !ok {
		t.Error("lodash should be in results")
	}

	// 验证其他 manifest 的依赖不在结果中
	if _, ok := result.UsageDetails["requests"]; ok {
		t.Error("requests (Python) should not be in npm results")
	}
	if _, ok := result.UsageDetails["gin-gonic"]; ok {
		t.Error("gin (Go) should not be in npm results")
	}
}

// TestIntegration_EmptyProject 测试空项目（无依赖）
func TestIntegration_EmptyProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "edge-empty")

	// 应该成功扫描
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// 应该没有依赖
	if result.TotalDeps != 0 {
		t.Errorf("Expected 0 total deps, got %d", result.TotalDeps)
	}
	if result.UsedDeps != 0 {
		t.Errorf("Expected 0 used deps, got %d", result.UsedDeps)
	}
	if result.UnusedDeps != 0 {
		t.Errorf("Expected 0 unused deps, got %d", result.UnusedDeps)
	}

	// UsageDetails 应该为空
	if len(result.UsageDetails) != 0 {
		t.Errorf("Expected empty UsageDetails, got %d entries", len(result.UsageDetails))
	}
}

// TestIntegration_ErrorScenarios 测试错误场景
func TestIntegration_ErrorScenarios(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		dir         string
		expectError bool
	}{
		{name: "invalid JSON", dir: "error-invalid-json", expectError: true},
		{name: "invalid TOML", dir: "error-invalid-toml", expectError: false},             // Cargo parser is lenient
		{name: "corrupted lockfile", dir: "error-corrupted-lockfile", expectError: false}, // Lockfile errors are ignored
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testdataDir := filepath.Join("..", "testdata", tt.dir)
			if _, err := os.Stat(testdataDir); os.IsNotExist(err) {
				t.Skipf("testdata not found: %s", testdataDir)
			}

			_, err := analyzer.Scan(testdataDir)

			if tt.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestIntegration_EdgeLargeProject 测试大量依赖的项目
func TestIntegration_EdgeLargeProject(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "edge-large")

	// 应该有大量依赖
	if result.TotalDeps < 100 {
		t.Errorf("Expected at least 100 deps, got %d", result.TotalDeps)
	}

	// 验证统计正确
	if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
		t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
			result.UsedDeps, result.UnusedDeps, result.TotalDeps)
	}

	// 验证 UsageDetails 包含所有依赖
	if len(result.UsageDetails) != result.TotalDeps {
		t.Errorf("UsageDetails count mismatch: got %d, expected %d",
			len(result.UsageDetails), result.TotalDeps)
	}
}

// TestIntegration_EdgeSpecialChars 测试包含特殊字符的依赖名
func TestIntegration_EdgeSpecialChars(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "edge-special-chars")

	// 应该成功扫描
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// 验证 scoped packages 被正确处理
	scopedPackages := []string{"@babel/core", "@babel/preset-env", "@types/node", "@typescript-eslint/parser"}
	for _, pkg := range scopedPackages {
		if _, ok := result.UsageDetails[pkg]; !ok {
			t.Errorf("Scoped package '%s' not found in results", pkg)
		}
	}

	// 验证统计正确
	if result.UsedDeps+result.UnusedDeps != result.TotalDeps {
		t.Errorf("Stats mismatch: Used(%d) + Unused(%d) != Total(%d)",
			result.UsedDeps, result.UnusedDeps, result.TotalDeps)
	}
}

// TestIntegration_NpmWorkspaces 测试 npm workspaces 项目
func TestIntegration_NpmWorkspaces(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "complex-npm-workspaces")
	helpers.AssertScanResult(t, result, "npm")

	// 应该只分析根目录的依赖（不递归分析 workspaces）
	if result.TotalDeps != 4 {
		t.Errorf("Expected 4 root deps, got %d", result.TotalDeps)
	}

	// 验证根目录的依赖
	expectedDeps := []string{"lodash", "axios", "typescript", "jest"}
	for _, dep := range expectedDeps {
		if _, ok := result.UsageDetails[dep]; !ok {
			t.Errorf("Expected dep '%s' not found in results", dep)
		}
	}
}

// TestIntegration_CargoWorkspaces 测试 Cargo workspaces 项目
func TestIntegration_CargoWorkspaces(t *testing.T) {
	t.Parallel()
	result := helpers.ScanTestdata(t, "complex-cargo-workspaces")
	helpers.AssertScanResult(t, result, "cargo")

	// 应该只分析根目录的依赖
	if result.TotalDeps != 2 {
		t.Errorf("Expected 2 root deps, got %d", result.TotalDeps)
	}

	// 验证根目录的依赖
	if _, ok := result.UsageDetails["serde"]; !ok {
		t.Error("serde not found in results")
	}
	if _, ok := result.UsageDetails["tokio"]; !ok {
		t.Error("tokio not found in results")
	}
}
