package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/config"
)

func TestScan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		setup          func(string) error
		expectedType   string
		expectedTotal  int
		expectedUsed   int
		expectedUnused int
		expectError    bool
	}{
		{
			name: "npm project with mixed usage",
			setup: func(dir string) error {
				pkgJSON := `{
					"name": "test-project",
					"dependencies": {
						"axios": "^1.0.0",
						"lodash": "^4.17.21",
						"unused-pkg": "^1.0.0"
					}
				}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
					return err
				}
				jsCode := `
import axios from 'axios';
import { debounce } from 'lodash';
`
				return os.WriteFile(filepath.Join(dir, "index.js"), []byte(jsCode), 0644)
			},
			expectedType:   "npm",
			expectedTotal:  3,
			expectedUsed:   2,
			expectedUnused: 1,
			expectError:    false,
		},
		{
			name: "go project with mixed usage",
			setup: func(dir string) error {
				goMod := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
	github.com/unused/pkg v1.0.0
)
`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
					return err
				}
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
				return os.WriteFile(filepath.Join(dir, "main.go"), []byte(goCode), 0644)
			},
			expectedType:   "go",
			expectedTotal:  3,
			expectedUsed:   2,
			expectedUnused: 1,
			expectError:    false,
		},
		{
			name: "rust project with mixed usage",
			setup: func(dir string) error {
				cargoToml := `[package]
name = "test-rust"
version = "0.1.0"
edition = "2021"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", features = ["full"] }
unused-crate = "1.0"
`
				if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
					return err
				}
				rustCode := `use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;

fn main() {
    println!("Hello");
}
`
				return os.WriteFile(filepath.Join(dir, "main.rs"), []byte(rustCode), 0644)
			},
			expectedType:   "cargo",
			expectedTotal:  3,
			expectedUsed:   2,
			expectedUnused: 1,
			expectError:    false,
		},
		{
			name: "python project with mixed usage",
			setup: func(dir string) error {
				requirements := `requests>=2.31.0
flask>=3.0.0
unused-package>=1.0.0
`
				if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte(requirements), 0644); err != nil {
					return err
				}
				pyCode := `import requests
from flask import Flask

app = Flask(__name__)
response = requests.get('https://api.example.com')
`
				return os.WriteFile(filepath.Join(dir, "app.py"), []byte(pyCode), 0644)
			},
			expectedType:   "pip",
			expectedTotal:  3,
			expectedUsed:   2,
			expectedUnused: 1,
			expectError:    false,
		},
		{
			name:           "empty project",
			setup:          func(dir string) error { return nil },
			expectedType:   "",
			expectedTotal:  0,
			expectedUsed:   0,
			expectedUnused: 0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			result, err := Scan(tmpDir)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Scan failed: %v", err)
			}

			if result.ManifestType != tt.expectedType {
				t.Errorf("Expected manifest type '%s', got '%s'", tt.expectedType, result.ManifestType)
			}

			if result.TotalDeps != tt.expectedTotal {
				t.Errorf("Expected %d total dependencies, got %d", tt.expectedTotal, result.TotalDeps)
			}

			if result.UsedDeps != tt.expectedUsed {
				t.Errorf("Expected %d used dependencies, got %d", tt.expectedUsed, result.UsedDeps)
			}

			if result.UnusedDeps != tt.expectedUnused {
				t.Errorf("Expected %d unused dependencies, got %d", tt.expectedUnused, result.UnusedDeps)
			}
		})
	}
}

func TestScanWithConfig(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建项目
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21",
			"@types/node": "^18.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建配置（忽略 @types/node）
	configContent := `ignore: ["@types/node"]`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	result, err := ScanWithConfig(tmpDir, cfg)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}

	// @types/node 应该被忽略，不计入总数
	if result.TotalDeps != 2 {
		t.Errorf("Expected 2 total deps (excluding ignored), got %d", result.TotalDeps)
	}

	if result.UsedDeps != 1 {
		t.Errorf("Expected 1 used dep, got %d", result.UsedDeps)
	}

	// @types/node 不应该出现在结果中
	if _, exists := result.UsageDetails["@types/node"]; exists {
		t.Error("@types/node should be ignored and not appear in results")
	}
}

func TestScanWithConfig_ExcludeDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建项目
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 在主目录创建使用 axios 的文件
	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 custom-dir 目录，在其中创建使用 lodash 的文件
	customDir := filepath.Join(tmpDir, "custom-dir")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	lodashCode := `import lodash from 'lodash';`
	if err := os.WriteFile(filepath.Join(customDir, "utils.js"), []byte(lodashCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 1: 不排除 custom-dir 时，lodash 应该被检测到
	cfg1 := &config.Config{
		ExcludeDirs: []string{},
		LockFile:    false,
	}
	result1, err := ScanWithConfig(tmpDir, cfg1)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if !result1.UsageDetails["lodash"].Used {
		t.Error("lodash should be used when custom-dir is not excluded")
	}

	// 测试 2: 排除 custom-dir 后，lodash 应该检测不到
	cfg2 := &config.Config{
		ExcludeDirs: []string{"custom-dir"},
		LockFile:    false,
	}
	result2, err := ScanWithConfig(tmpDir, cfg2)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if result2.UsageDetails["lodash"].Used {
		t.Error("lodash should NOT be used when custom-dir is excluded")
	}
}

func TestScanWithConfig_ExcludeFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建项目
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 在主目录创建使用 axios 的文件
	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建使用 lodash 的测试文件
	lodashCode := `import lodash from 'lodash';`
	if err := os.WriteFile(filepath.Join(tmpDir, "utils.test.js"), []byte(lodashCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 1: 不排除测试文件时，lodash 应该被检测到
	cfg1 := &config.Config{
		ExcludeFiles: []string{},
		LockFile:     false,
	}
	result1, err := ScanWithConfig(tmpDir, cfg1)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if !result1.UsageDetails["lodash"].Used {
		t.Error("lodash should be used when test files are not excluded")
	}

	// 测试 2: 排除 *.test.js 文件后，lodash 应该检测不到
	cfg2 := &config.Config{
		ExcludeFiles: []string{"*.test.js"},
		LockFile:     false,
	}
	result2, err := ScanWithConfig(tmpDir, cfg2)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if result2.UsageDetails["lodash"].Used {
		t.Error("lodash should NOT be used when *.test.js files are excluded")
	}
}

func TestScanWithConfig_ReadNodeModules(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建项目
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 在主目录创建使用 axios 的文件
	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 node_modules 目录，在其中创建使用 lodash 的文件
	nodeModulesDir := filepath.Join(tmpDir, "node_modules", "some-pkg")
	if err := os.MkdirAll(nodeModulesDir, 0755); err != nil {
		t.Fatal(err)
	}
	lodashCode := `import lodash from 'lodash';`
	if err := os.WriteFile(filepath.Join(nodeModulesDir, "index.js"), []byte(lodashCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 1: 不读取 node_modules 时，lodash 应该检测不到
	cfg1 := &config.Config{
		ReadNodeModules: false,
		LockFile:        false,
	}
	result1, err := ScanWithConfig(tmpDir, cfg1)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if result1.UsageDetails["lodash"].Used {
		t.Error("lodash should NOT be used when node_modules is not read")
	}

	// 测试 2: 读取 node_modules 后，lodash 应该被检测到
	cfg2 := &config.Config{
		ReadNodeModules: true,
		LockFile:        false,
	}
	result2, err := ScanWithConfig(tmpDir, cfg2)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if !result2.UsageDetails["lodash"].Used {
		t.Error("lodash should be used when node_modules is read")
	}
}

func TestScanWithConfig_LockFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建项目
	pkgJSON := `{
		"name": "test-project",
		"dependencies": {
			"axios": "^1.0.0"
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建使用 axios 的文件
	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 package-lock.json，包含间接依赖
	lockJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"lockfileVersion": 2,
		"packages": {
			"": {
				"name": "test-project",
				"version": "1.0.0",
				"dependencies": {
					"axios": "^1.0.0"
				}
			},
			"node_modules/axios": {
				"version": "1.6.0",
				"resolved": "https://registry.npmjs.org/axios/-/axios-1.6.0.tgz",
				"dependencies": {
					"follow-redirects": "^1.15.0"
				}
			},
			"node_modules/follow-redirects": {
				"version": "1.15.3",
				"resolved": "https://registry.npmjs.org/follow-redirects/-/follow-redirects-1.15.3.tgz"
			}
		}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(lockJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 1: 不启用 lockfile 分析时，IndirectDeps 应该为空
	cfg1 := &config.Config{
		LockFile: false,
	}
	result1, err := ScanWithConfig(tmpDir, cfg1)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if len(result1.IndirectDeps) > 0 {
		t.Errorf("IndirectDeps should be empty when lockfile is disabled, got %v", result1.IndirectDeps)
	}

	// 测试 2: 启用 lockfile 分析后，应该检测到间接依赖
	cfg2 := &config.Config{
		LockFile: true,
	}
	result2, err := ScanWithConfig(tmpDir, cfg2)
	if err != nil {
		t.Fatalf("ScanWithConfig failed: %v", err)
	}
	if len(result2.IndirectDeps) == 0 {
		t.Error("IndirectDeps should not be empty when lockfile is enabled")
	}

	// 验证间接依赖包含 follow-redirects
	found := false
	for _, dep := range result2.IndirectDeps {
		if dep == "follow-redirects" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("IndirectDeps should contain 'follow-redirects', got %v", result2.IndirectDeps)
	}
}
