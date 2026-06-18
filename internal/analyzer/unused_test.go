package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanNpmProject(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 package.json
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21",
			"unused-pkg": "^1.0.0"
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建源文件
	jsCode := `
import axios from 'axios';
import { debounce } from 'lodash';
`

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "npm" {
		t.Errorf("Expected manifest type 'npm', got '%s'", result.ManifestType)
	}

	if result.TotalDeps != 3 {
		t.Errorf("Expected 3 total dependencies, got %d", result.TotalDeps)
	}

	if result.UsedDeps != 2 {
		t.Errorf("Expected 2 used dependencies, got %d", result.UsedDeps)
	}

	if result.UnusedDeps != 1 {
		t.Errorf("Expected 1 unused dependency, got %d", result.UnusedDeps)
	}

	// 验证未使用的依赖
	if !result.UsageDetails["unused-pkg"].Used {
		// 正确，unused-pkg 应该未被使用
	} else {
		t.Error("Expected 'unused-pkg' to be unused")
	}

	if !result.UsageDetails["axios"].Used {
		t.Error("Expected 'axios' to be used")
	}

	if !result.UsageDetails["lodash"].Used {
		t.Error("Expected 'lodash' to be used")
	}
}

func TestScanGoProject(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建 go.mod
	goMod := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
	github.com/unused/pkg v1.0.0
)
`

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建源文件
	goCode := `package main

import (
	"github.com/gin-gonic/gin"
	alias "github.com/spf13/cobra"
)

func main() {
	_ = gin.Default()
	_ = alias.Command{}
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.ManifestType != "go" {
		t.Errorf("Expected manifest type 'go', got '%s'", result.ManifestType)
	}

	if result.TotalDeps != 3 {
		t.Errorf("Expected 3 total dependencies, got %d", result.TotalDeps)
	}

	if result.UsedDeps != 2 {
		t.Errorf("Expected 2 used dependency, got %d", result.UsedDeps)
	}

	if result.UnusedDeps != 1 {
		t.Errorf("Expected 1 unused dependency, got %d", result.UnusedDeps)
	}
}

func TestScanNoProject(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Scan(tmpDir)
	if err == nil {
		t.Error("Expected error for missing project, got nil")
	}
}

func TestScanEmptyProject(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建空的 package.json
	pkgJSON := `{"name": "empty", "dependencies": {}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Scan(tmpDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.TotalDeps != 0 {
		t.Errorf("Expected 0 dependencies, got %d", result.TotalDeps)
	}
}
