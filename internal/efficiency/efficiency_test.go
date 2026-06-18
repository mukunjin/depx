package efficiency

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEfficiencyAnalysis 效率分析综合测试（table-driven）
func TestEfficiencyAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		deps     []string
		expected map[string]struct {
			usedExports []string
			efficiency  float64
		}
		expectError bool
	}{
		{
			name: "single file with multiple imports",
			files: map[string]string{
				"index.js": `import { debounce } from "lodash";
import { useState, useEffect } from "react";
import axios from "axios";
import * as path from "path";
`,
			},
			deps: []string{"lodash", "react", "axios", "path"},
			expected: map[string]struct {
				usedExports []string
				efficiency  float64
			}{
				"lodash": {usedExports: []string{"debounce"}, efficiency: 3.3},
				"react":  {usedExports: []string{"useEffect", "useState"}, efficiency: 6.7},
				"axios":  {usedExports: []string{"default"}, efficiency: 3.3},
				"path":   {usedExports: []string{"*"}, efficiency: 3.3},
			},
			expectError: false,
		},
		{
			name: "multiple files with deduplication",
			files: map[string]string{
				"file1.js": `import { debounce } from "lodash";`,
				"file2.js": `import { throttle } from "lodash";`,
			},
			deps: []string{"lodash"},
			expected: map[string]struct {
				usedExports []string
				efficiency  float64
			}{
				"lodash": {usedExports: []string{"debounce", "throttle"}, efficiency: 6.7},
			},
			expectError: false,
		},
		{
			name: "no matching dependencies",
			files: map[string]string{
				"index.js": `import { something } from "other-package";`,
			},
			deps: []string{"lodash", "react"},
			expected: map[string]struct {
				usedExports []string
				efficiency  float64
			}{},
			expectError: false,
		},
		{
			name:  "empty project",
			files: map[string]string{},
			deps:  []string{"lodash"},
			expected: map[string]struct {
				usedExports []string
				efficiency  float64
			}{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// 创建测试文件
			for filename, content := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			results, err := AnalyzeEfficiency(tmpDir, tt.deps)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("AnalyzeEfficiency failed: %v", err)
			}

			// 验证结果
			for pkg, exp := range tt.expected {
				result, ok := results[pkg]
				if !ok {
					t.Errorf("Package %s not found in results", pkg)
					continue
				}

				// 验证使用的导出
				if len(result.UsedExports) != len(exp.usedExports) {
					t.Errorf("Package %s: expected %d exports, got %d: %v",
						pkg, len(exp.usedExports), len(result.UsedExports), result.UsedExports)
				} else {
					for i, export := range exp.usedExports {
						if result.UsedExports[i] != export {
							t.Errorf("Package %s: export[%d] = %s, expected %s",
								pkg, i, result.UsedExports[i], export)
						}
					}
				}

				// 验证效率百分比（允许 0.1% 误差）
				if result.Efficiency < exp.efficiency-0.1 || result.Efficiency > exp.efficiency+0.1 {
					t.Errorf("Package %s: efficiency = %.1f%%, expected %.1f%%",
						pkg, result.Efficiency, exp.efficiency)
				}
			}
		})
	}
}

// TestExportAnalysis 导出分析测试（table-driven）
func TestExportAnalysis(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "named exports",
			code:     `export { foo, bar };`,
			expected: []string{"foo", "bar"},
		},
		{
			name: "function exports",
			code: `export function baz() {}
export function qux() {}`,
			expected: []string{"baz", "qux"},
		},
		{
			name: "const exports",
			code: `export const PI = 3.14;
export const VERSION = "1.0.0";`,
			expected: []string{"PI", "VERSION"},
		},
		{
			name: "class exports",
			code: `export class MyClass {}
export class AnotherClass {}`,
			expected: []string{"MyClass", "AnotherClass"},
		},
		{
			name: "mixed exports",
			code: `export { foo, bar };
export function baz() {}
export const qux = 42;
export class MyClass {}`,
			expected: []string{"foo", "bar", "baz", "qux", "MyClass"},
		},
		{
			name:     "no exports",
			code:     `const x = 1; function y() {}`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "module.js")
			if err := os.WriteFile(filePath, []byte(tt.code), 0644); err != nil {
				t.Fatal(err)
			}

			exports, err := AnalyzeExports(filePath)
			if err != nil {
				t.Fatalf("AnalyzeExports failed: %v", err)
			}

			if len(exports) != len(tt.expected) {
				t.Errorf("Expected %d exports, got %d: %v", len(tt.expected), len(exports), exports)
				return
			}

			// 验证所有预期导出都存在
			exportMap := make(map[string]bool)
			for _, exp := range exports {
				exportMap[exp] = true
			}

			for _, exp := range tt.expected {
				if !exportMap[exp] {
					t.Errorf("Expected export %s not found", exp)
				}
			}
		})
	}
}

// TestEstimateTotalExports 估算总导出数测试（table-driven）
func TestEstimateTotalExports(t *testing.T) {
	tests := []struct {
		name     string
		used     int
		expected int
	}{
		{"1 export", 1, 30},
		{"2 exports", 2, 30},
		{"3 exports", 3, 50},
		{"5 exports", 5, 50},
		{"6 exports", 6, 80},
		{"10 exports", 10, 80},
		{"11 exports", 11, 120},
		{"20 exports", 20, 120},
		{"21 exports", 21, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTotalExports(tt.used)
			if result != tt.expected {
				t.Errorf("estimateTotalExports(%d) = %d, expected %d", tt.used, result, tt.expected)
			}
		})
	}
}

// BenchmarkAnalyzeEfficiency 效率分析性能基准测试
func BenchmarkAnalyzeEfficiency(b *testing.B) {
	tmpDir := b.TempDir()

	// 创建大型测试文件
	code := `import { debounce, throttle, clone, merge } from "lodash";
import { useState, useEffect, useCallback, useMemo } from "react";
import axios from "axios";
import * as path from "path";
import { format, parse, isValid } from "date-fns";
`

	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+".js")
		if err := os.WriteFile(filename, []byte(code), 0644); err != nil {
			b.Fatal(err)
		}
	}

	deps := []string{"lodash", "react", "axios", "path", "date-fns"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AnalyzeEfficiency(tmpDir, deps)
	}
}

// BenchmarkAnalyzeExports 导出分析性能基准测试
func BenchmarkAnalyzeExports(b *testing.B) {
	tmpDir := b.TempDir()

	code := `export { foo, bar, baz };
export function func1() {}
export function func2() {}
export const CONST1 = 1;
export const CONST2 = 2;
export class Class1 {}
export class Class2 {}
`

	filePath := filepath.Join(tmpDir, "module.js")
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = AnalyzeExports(filePath)
	}
}
