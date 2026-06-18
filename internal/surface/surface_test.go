package surface

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeSurface(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个测试文件
	file1 := `import axios from "axios";
import { debounce } from "lodash";
`
	file2 := `import axios from "axios";
import moment from "moment";
`
	file3 := `import { throttle } from "lodash";
import "lodash/fp";
`

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.js"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.js"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file3.js"), []byte(file3), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "lodash", "moment"}
	results, err := AnalyzeSurface(tmpDir, deps)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 axios
	if result, ok := results["axios"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for axios, got %d", len(result.Files))
		}
		if result.RefCount != 2 {
			t.Errorf("Expected 2 refs for axios, got %d", result.RefCount)
		}
		if result.Criticality != "Low" {
			t.Errorf("Expected Low criticality for axios, got %s", result.Criticality)
		}
	} else {
		t.Error("axios not found in results")
	}

	// 检查 lodash
	if result, ok := results["lodash"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for lodash, got %d", len(result.Files))
		}
		// 应该有 2 个模块：lodash 和 lodash/fp
		if len(result.Modules) != 2 {
			t.Errorf("Expected 2 modules for lodash, got %d: %v", len(result.Modules), result.Modules)
		}
		if result.RefCount != 3 {
			t.Errorf("Expected 3 refs for lodash, got %d", result.RefCount)
		}
	} else {
		t.Error("lodash not found in results")
	}

	// 检查 moment
	if result, ok := results["moment"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for moment, got %d", len(result.Files))
		}
		if result.RefCount != 1 {
			t.Errorf("Expected 1 ref for moment, got %d", result.RefCount)
		}
	} else {
		t.Error("moment not found in results")
	}
}

func TestAnalyzeSurfaceWithRequire(t *testing.T) {
	tmpDir := t.TempDir()

	code := `const axios = require("axios");
const lodash = require("lodash");
`

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "lodash"}
	results, err := AnalyzeSurface(tmpDir, deps)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if result, ok := results["axios"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for axios, got %d", len(result.Files))
		}
	} else {
		t.Error("axios not found in results")
	}
}

func TestAnalyzeSurfaceScopedPackage(t *testing.T) {
	tmpDir := t.TempDir()

	code := `import { something } from "@org/pkg";
import { other } from "@org/pkg/sub";
`

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"@org/pkg"}
	results, err := AnalyzeSurface(tmpDir, deps)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if result, ok := results["@org/pkg"]; ok {
		if len(result.Modules) != 2 {
			t.Errorf("Expected 2 modules for @org/pkg, got %d: %v", len(result.Modules), result.Modules)
		}
	} else {
		t.Error("@org/pkg not found in results")
	}
}

func TestResolvePackageName(t *testing.T) {
	tests := []struct {
		importPath string
		expected   string
	}{
		{"axios", "axios"},
		{"lodash/debounce", "lodash"},
		{"@org/pkg", "@org/pkg"},
		{"@org/pkg/sub", "@org/pkg"},
		{"./relative", ""},
		{"../parent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			result := resolvePackageName(tt.importPath)
			if result != tt.expected {
				t.Errorf("resolvePackageName(%s) = %s, expected %s", tt.importPath, result, tt.expected)
			}
		})
	}
}

func TestCalculateCriticality(t *testing.T) {
	tests := []struct {
		name      string
		fileCount int
		refCount  int
		expected  string
	}{
		{"Low - 1 file, 1 ref", 1, 1, "Low"},
		{"Low - 2 files, 5 refs", 2, 5, "Low"},
		{"Medium - 3 files, 10 refs", 3, 10, "Medium"},
		{"Medium - 9 files, 49 refs", 9, 49, "Medium"},
		{"High - 10 files, 50 refs", 10, 50, "High"},
		{"High - 20 files, 100 refs", 20, 100, "High"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SurfaceResult{
				Files:    make([]string, tt.fileCount),
				RefCount: tt.refCount,
			}
			criticality := calculateCriticality(result)
			if criticality != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, criticality)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	code := `import axios from "axios";
import { debounce } from "lodash";
const fs = require("fs");
import("./dynamic");
`

	imports := extractImports(code)

	expected := map[string]bool{
		"axios":     true,
		"lodash":    true,
		"fs":        true,
		"./dynamic": true,
	}

	if len(imports) != len(expected) {
		t.Errorf("Expected %d imports, got %d: %v", len(expected), len(imports), imports)
	}

	for _, imp := range imports {
		if !expected[imp] {
			t.Errorf("Unexpected import: %s", imp)
		}
	}
}

func TestSurfaceResultDedup(t *testing.T) {
	// 测试 map 去重逻辑
	result := &SurfaceResult{
		Package:   "test-pkg",
		Files:     []string{},
		Modules:   []string{},
		fileSet:   make(map[string]struct{}),
		moduleSet: make(map[string]struct{}),
	}

	// 添加相同文件两次
	relPath := "test.js"
	if _, exists := result.fileSet[relPath]; !exists {
		result.fileSet[relPath] = struct{}{}
		result.Files = append(result.Files, relPath)
	}
	if _, exists := result.fileSet[relPath]; !exists {
		result.fileSet[relPath] = struct{}{}
		result.Files = append(result.Files, relPath)
	}

	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}
}
