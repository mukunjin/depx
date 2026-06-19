package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/config"
)

func TestScan(t *testing.T) {
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
