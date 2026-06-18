package usage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/depx/depx/internal/manifest"
)

// TestBoundary_EmptyFile 测试空文件处理
func TestBoundary_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建空的 JS 文件
	if err := os.WriteFile(filepath.Join(tmpDir, "empty.js"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建空的 Go 文件
	if err := os.WriteFile(filepath.Join(tmpDir, "empty.go"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 JS 分析器
	jsAnalyzer := NewJSAnalyzer()
	jsResults, err := jsAnalyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatalf("JS analyzer failed on empty file: %v", err)
	}
	if jsResults["axios"].Used {
		t.Error("Empty file should not mark dependency as used")
	}

	// 测试 Go 分析器
	goAnalyzer := NewGoAnalyzer()
	goResults, err := goAnalyzer.Analyze(tmpDir, []string{"github.com/gin-gonic/gin"})
	if err != nil {
		t.Fatalf("Go analyzer failed on empty file: %v", err)
	}
	if goResults["github.com/gin-gonic/gin"].Used {
		t.Error("Empty file should not mark dependency as used")
	}
}

// TestBoundary_LargeFile 测试大文件处理
func TestBoundary_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建一个约 1MB 的文件
	var sb strings.Builder
	sb.WriteString("import axios from 'axios';\n")
	for i := 0; i < 10000; i++ {
		sb.WriteString("// This is a comment line to make the file large\n")
		sb.WriteString("const x = 1;\n")
	}
	sb.WriteString("import lodash from 'lodash';\n")

	if err := os.WriteFile(filepath.Join(tmpDir, "large.js"), []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios", "lodash"})
	if err != nil {
		t.Fatalf("Failed to analyze large file: %v", err)
	}

	if !results["axios"].Used {
		t.Error("Should detect axios in large file")
	}
	if !results["lodash"].Used {
		t.Error("Should detect lodash in large file")
	}
}

// TestBoundary_Symlinks 测试符号链接处理
func TestBoundary_Symlinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink test skipped on Windows")
	}

	tmpDir := t.TempDir()

	// 创建真实目录和文件
	realDir := filepath.Join(tmpDir, "real")
	if err := os.MkdirAll(realDir, 0755); err != nil {
		t.Fatal(err)
	}

	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(realDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建符号链接
	linkDir := filepath.Join(tmpDir, "link")
	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("Cannot create symlink: %v", err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatalf("Failed to analyze with symlinks: %v", err)
	}

	// 应该至少检测到一次
	if !results["axios"].Used {
		t.Error("Should detect axios through symlink")
	}
}

// TestBoundary_SpecialCharacters 测试特殊字符路径
func TestBoundary_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建包含特殊字符的目录
	specialDir := filepath.Join(tmpDir, "test-dir_with.special+chars")
	if err := os.MkdirAll(specialDir, 0755); err != nil {
		t.Fatal(err)
	}

	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(specialDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatalf("Failed with special characters in path: %v", err)
	}

	if !results["axios"].Used {
		t.Error("Should handle special characters in path")
	}
}

// TestBoundary_NonExistentDirectory 测试不存在的目录
func TestBoundary_NonExistentDirectory(t *testing.T) {
	analyzer := NewJSAnalyzer()
	_, err := analyzer.Analyze("/non/existent/directory", []string{"axios"})
	if err == nil {
		t.Error("Should return error for non-existent directory")
	}
}

// TestBoundary_UnreadableFile 测试不可读文件
func TestBoundary_UnreadableFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission test skipped on Windows")
	}

	tmpDir := t.TempDir()

	// 创建文件
	filePath := filepath.Join(tmpDir, "unreadable.js")
	if err := os.WriteFile(filePath, []byte("import axios from 'axios';"), 0644); err != nil {
		t.Fatal(err)
	}

	// 移除读权限
	if err := os.Chmod(filePath, 0000); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(filePath, 0644) // 恢复权限以便清理

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatalf("Should not fail on unreadable files: %v", err)
	}

	// 应该跳过不可读文件
	if results["axios"].Used {
		t.Error("Should not detect import in unreadable file")
	}
}

// TestBoundary_MultipleExtensions 测试多扩展名文件
func TestBoundary_MultipleExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建各种扩展名的文件
	files := map[string]string{
		"file.min.js":  "import axios from 'axios';",
		"file.test.js": "import lodash from 'lodash';",
		"file.spec.ts": "import react from 'react';",
		"file.d.ts":    "export declare const x: string;",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "react"}
	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Failed to analyze multiple extensions: %v", err)
	}

	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Should detect %s in file with multiple extensions", dep)
		}
	}
}

// TestBoundary_DuplicateImports 测试重复导入
func TestBoundary_DuplicateImports(t *testing.T) {
	tmpDir := t.TempDir()

	// 同一文件中多次导入同一包
	jsCode := `
import axios from 'axios';
import axios2 from 'axios';
const { get } = require('axios');
`
	if err := os.WriteFile(filepath.Join(tmpDir, "duplicate.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatal(err)
	}

	if !results["axios"].Used {
		t.Error("Should detect axios")
	}

	// RefCount 应该反映实际引用次数
	if results["axios"].RefCount < 3 {
		t.Errorf("Expected RefCount >= 3, got %d", results["axios"].RefCount)
	}

	// UsedIn 应该只包含一次文件路径
	if len(results["axios"].UsedIn) != 1 {
		t.Errorf("Expected 1 file in UsedIn, got %d", len(results["axios"].UsedIn))
	}
}

// TestBoundary_NestedDirectories 测试深层嵌套目录
func TestBoundary_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 10 层深的目录
	deepDir := tmpDir
	for i := 0; i < 10; i++ {
		deepDir = filepath.Join(deepDir, "level"+string(rune('0'+i)))
	}
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatal(err)
	}

	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(deepDir, "deep.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatal(err)
	}

	if !results["axios"].Used {
		t.Error("Should detect axios in deeply nested directory")
	}
}

// TestBoundary_MixedLineEndings 测试混合换行符
func TestBoundary_MixedLineEndings(t *testing.T) {
	tmpDir := t.TempDir()

	// 混合使用 \n 和 \r\n
	jsCode := "import axios from 'axios';\r\nimport lodash from 'lodash';\nimport react from 'react';\r"
	if err := os.WriteFile(filepath.Join(tmpDir, "mixed.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "react"}
	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Should detect %s with mixed line endings", dep)
		}
	}
}

// TestBoundary_EmptyDependencies 测试空依赖列表
func TestBoundary_EmptyDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{})
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 0 {
		t.Error("Should return empty results for empty dependencies")
	}
}

// TestBoundary_CommentsInImports 测试注释中的导入语句
func TestBoundary_CommentsInImports(t *testing.T) {
	tmpDir := t.TempDir()

	jsCode := `
import axios from 'axios';
// import lodash from 'lodash';
/* 
import react from 'react';
*/
import('dynamic');
`
	if err := os.WriteFile(filepath.Join(tmpDir, "comments.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "react", "dynamic"}
	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	if !results["axios"].Used {
		t.Error("Should detect axios (not in comment)")
	}
	if results["lodash"].Used {
		t.Error("Should not detect lodash (in single-line comment)")
	}
	if results["react"].Used {
		t.Error("Should not detect react (in multi-line comment)")
	}
	if !results["dynamic"].Used {
		t.Error("Should detect dynamic import")
	}
}

// TestBoundary_UnicodeContent 测试 Unicode 内容
func TestBoundary_UnicodeContent(t *testing.T) {
	tmpDir := t.TempDir()

	// 包含中文注释和变量名
	jsCode := `
// 这是中文注释
import axios from 'axios';
const 数据 = axios.get('/api');
`
	if err := os.WriteFile(filepath.Join(tmpDir, "unicode.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"})
	if err != nil {
		t.Fatal(err)
	}

	if !results["axios"].Used {
		t.Error("Should handle Unicode content correctly")
	}
}

// TestUsageResult_Fields 测试 UsageResult 字段完整性
func TestUsageResult_Fields(t *testing.T) {
	tmpDir := t.TempDir()

	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios", "unused"})
	if err != nil {
		t.Fatal(err)
	}

	// 测试已使用的依赖
	axiosResult := results["axios"]
	if axiosResult == nil {
		t.Fatal("axios result should not be nil")
	}
	if !axiosResult.Used {
		t.Error("axios should be marked as used")
	}
	if axiosResult.Package != "axios" {
		t.Errorf("Expected package name 'axios', got '%s'", axiosResult.Package)
	}
	if len(axiosResult.UsedIn) == 0 {
		t.Error("UsedIn should not be empty for used dependency")
	}
	if axiosResult.RefCount == 0 {
		t.Error("RefCount should not be 0 for used dependency")
	}

	// 测试未使用的依赖
	unusedResult := results["unused"]
	if unusedResult == nil {
		t.Fatal("unused result should not be nil")
	}
	if unusedResult.Used {
		t.Error("unused should not be marked as used")
	}
	if len(unusedResult.UsedIn) != 0 {
		t.Error("UsedIn should be empty for unused dependency")
	}
	if unusedResult.RefCount != 0 {
		t.Error("RefCount should be 0 for unused dependency")
	}
}

// TestManifest_UsageResult_Integration 测试 manifest 和 usage 的集成
func TestManifest_UsageResult_Integration(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 package.json
	pkgJSON := `{
		"name": "test",
		"dependencies": {
			"axios": "^1.0.0",
			"unused": "^1.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建源文件
	jsCode := "import axios from 'axios';\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 manifest
	m, err := manifest.NewNpmManifest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	deps, err := m.Dependencies()
	if err != nil {
		t.Fatal(err)
	}

	// 测试 usage
	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatal(err)
	}

	// 验证结果
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	usedCount := 0
	for _, r := range results {
		if r.Used {
			usedCount++
		}
	}

	if usedCount != 1 {
		t.Errorf("Expected 1 used dependency, got %d", usedCount)
	}
}
