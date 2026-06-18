package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectWorkspaces(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(string) error
		expectedCount int
		expectError   bool
	}{
		{
			name: "npm workspaces with glob pattern",
			setup: func(dir string) error {
				// 创建根 package.json
				rootPkg := map[string]interface{}{
					"name":       "monorepo-root",
					"version":    "1.0.0",
					"workspaces": []string{"packages/*"},
				}
				rootData, _ := json.Marshal(rootPkg)
				if err := os.WriteFile(filepath.Join(dir, "package.json"), rootData, 0644); err != nil {
					return err
				}

				// 创建工作区目录
				pkgADir := filepath.Join(dir, "packages", "pkg-a")
				pkgBDir := filepath.Join(dir, "packages", "pkg-b")
				if err := os.MkdirAll(pkgADir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(pkgBDir, 0755); err != nil {
					return err
				}

				// 创建 pkg-a 的 package.json
				pkgA := map[string]interface{}{
					"name":    "pkg-a",
					"version": "1.0.0",
					"dependencies": map[string]string{
						"lodash": "^4.17.21",
					},
				}
				pkgAData, _ := json.Marshal(pkgA)
				if err := os.WriteFile(filepath.Join(pkgADir, "package.json"), pkgAData, 0644); err != nil {
					return err
				}

				// 创建 pkg-b 的 package.json
				pkgB := map[string]interface{}{
					"name":    "pkg-b",
					"version": "1.0.0",
					"dependencies": map[string]string{
						"axios": "^1.6.0",
					},
				}
				pkgBData, _ := json.Marshal(pkgB)
				if err := os.WriteFile(filepath.Join(pkgBDir, "package.json"), pkgBData, 0644); err != nil {
					return err
				}

				// 创建源文件
				if err := os.WriteFile(filepath.Join(pkgADir, "index.js"), []byte(`import _ from 'lodash';`), 0644); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(pkgBDir, "index.js"), []byte(`import axios from 'axios';`), 0644)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "yarn workspaces with explicit paths",
			setup: func(dir string) error {
				rootPkg := map[string]interface{}{
					"name":       "monorepo-root",
					"version":    "1.0.0",
					"workspaces": []string{"frontend", "backend"},
				}
				rootData, _ := json.Marshal(rootPkg)
				if err := os.WriteFile(filepath.Join(dir, "package.json"), rootData, 0644); err != nil {
					return err
				}

				frontendDir := filepath.Join(dir, "frontend")
				backendDir := filepath.Join(dir, "backend")
				if err := os.MkdirAll(frontendDir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(backendDir, 0755); err != nil {
					return err
				}

				frontend := map[string]interface{}{
					"name":         "frontend",
					"version":      "1.0.0",
					"dependencies": map[string]string{"react": "^18.0.0"},
				}
				frontendData, _ := json.Marshal(frontend)
				if err := os.WriteFile(filepath.Join(frontendDir, "package.json"), frontendData, 0644); err != nil {
					return err
				}

				backend := map[string]interface{}{
					"name":         "backend",
					"version":      "1.0.0",
					"dependencies": map[string]string{"express": "^4.18.0"},
				}
				backendData, _ := json.Marshal(backend)
				return os.WriteFile(filepath.Join(backendDir, "package.json"), backendData, 0644)
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "no workspaces defined",
			setup: func(dir string) error {
				rootPkg := map[string]interface{}{
					"name":    "single-project",
					"version": "1.0.0",
				}
				rootData, _ := json.Marshal(rootPkg)
				return os.WriteFile(filepath.Join(dir, "package.json"), rootData, 0644)
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "no package.json",
			setup: func(dir string) error {
				return nil
			},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			workspaces, err := detectWorkspaces(tmpDir)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("detectWorkspaces failed: %v", err)
			}

			if len(workspaces) != tt.expectedCount {
				t.Errorf("Expected %d workspaces, got %d", tt.expectedCount, len(workspaces))
			}
		})
	}
}

func TestScanMonorepo(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建根 package.json
	rootPkg := map[string]interface{}{
		"name":       "monorepo-root",
		"version":    "1.0.0",
		"workspaces": []string{"packages/*"},
	}
	rootData, _ := json.Marshal(rootPkg)
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), rootData, 0644); err != nil {
		t.Fatal(err)
	}

	// 创建工作区目录
	pkgADir := filepath.Join(tmpDir, "packages", "pkg-a")
	pkgBDir := filepath.Join(tmpDir, "packages", "pkg-b")
	if err := os.MkdirAll(pkgADir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(pkgBDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 创建 pkg-a 的 package.json
	pkgA := map[string]interface{}{
		"name":    "pkg-a",
		"version": "1.0.0",
		"dependencies": map[string]string{
			"lodash": "^4.17.21",
			"unused": "^1.0.0",
		},
	}
	pkgAData, _ := json.Marshal(pkgA)
	if err := os.WriteFile(filepath.Join(pkgADir, "package.json"), pkgAData, 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 pkg-b 的 package.json
	pkgB := map[string]interface{}{
		"name":    "pkg-b",
		"version": "1.0.0",
		"dependencies": map[string]string{
			"axios": "^1.6.0",
		},
	}
	pkgBData, _ := json.Marshal(pkgB)
	if err := os.WriteFile(filepath.Join(pkgBDir, "package.json"), pkgBData, 0644); err != nil {
		t.Fatal(err)
	}

	// 创建源文件
	if err := os.WriteFile(filepath.Join(pkgADir, "index.js"), []byte(`import _ from 'lodash';`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgBDir, "index.js"), []byte(`import axios from 'axios';`), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试扫描 monorepo
	result, err := ScanMonorepo(tmpDir)
	if err != nil {
		t.Fatalf("ScanMonorepo failed: %v", err)
	}

	if result.Path != tmpDir {
		t.Errorf("Expected path %s, got %s", tmpDir, result.Path)
	}

	if len(result.Workspaces) != 2 {
		t.Errorf("Expected 2 workspaces, got %d", len(result.Workspaces))
	}

	if len(result.WorkspaceResults) != 2 {
		t.Errorf("Expected 2 workspace results, got %d", len(result.WorkspaceResults))
	}

	// 检查依赖统计
	if result.TotalDeps != 3 {
		t.Errorf("Expected 3 total deps, got %d", result.TotalDeps)
	}

	if result.UsedDeps != 2 {
		t.Errorf("Expected 2 used deps, got %d", result.UsedDeps)
	}

	if result.UnusedDeps != 1 {
		t.Errorf("Expected 1 unused dep, got %d", result.UnusedDeps)
	}
}

func TestExpandWorkspacePatterns(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(string) error
		patterns      []string
		expectedCount int
		expectError   bool
	}{
		{
			name: "glob pattern packages/*",
			setup: func(dir string) error {
				pkgADir := filepath.Join(dir, "packages", "pkg-a")
				pkgBDir := filepath.Join(dir, "packages", "pkg-b")
				otherDir := filepath.Join(dir, "other")
				if err := os.MkdirAll(pkgADir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(pkgBDir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(otherDir, 0755); err != nil {
					return err
				}

				for _, d := range []string{pkgADir, pkgBDir} {
					pkg := map[string]interface{}{
						"name":    filepath.Base(d),
						"version": "1.0.0",
					}
					data, _ := json.Marshal(pkg)
					if err := os.WriteFile(filepath.Join(d, "package.json"), data, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			patterns:      []string{"packages/*"},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "explicit paths",
			setup: func(dir string) error {
				frontendDir := filepath.Join(dir, "frontend")
				backendDir := filepath.Join(dir, "backend")
				if err := os.MkdirAll(frontendDir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(backendDir, 0755); err != nil {
					return err
				}

				for _, d := range []string{frontendDir, backendDir} {
					pkg := map[string]interface{}{
						"name":    filepath.Base(d),
						"version": "1.0.0",
					}
					data, _ := json.Marshal(pkg)
					if err := os.WriteFile(filepath.Join(d, "package.json"), data, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			patterns:      []string{"frontend", "backend"},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "mixed patterns",
			setup: func(dir string) error {
				pkgADir := filepath.Join(dir, "packages", "pkg-a")
				toolsDir := filepath.Join(dir, "tools")
				if err := os.MkdirAll(pkgADir, 0755); err != nil {
					return err
				}
				if err := os.MkdirAll(toolsDir, 0755); err != nil {
					return err
				}

				for _, d := range []string{pkgADir, toolsDir} {
					pkg := map[string]interface{}{
						"name":    filepath.Base(d),
						"version": "1.0.0",
					}
					data, _ := json.Marshal(pkg)
					if err := os.WriteFile(filepath.Join(d, "package.json"), data, 0644); err != nil {
						return err
					}
				}
				return nil
			},
			patterns:      []string{"packages/*", "tools"},
			expectedCount: 2,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			workspaces, err := expandWorkspacePatterns(tmpDir, tt.patterns)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expandWorkspacePatterns failed: %v", err)
			}

			if len(workspaces) != tt.expectedCount {
				t.Errorf("Expected %d workspaces, got %d", tt.expectedCount, len(workspaces))
			}
		})
	}
}
