package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestRunSurface(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(string) error
		path         string
		configPath   string
		showIndirect bool
		expectError  bool
	}{
		{
			name: "surface analysis",
			setup: func(dir string) error {
				// 创建 package.json
				pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21", "axios": "^1.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				// 创建源文件
				srcContent := `import { debounce } from "lodash";
import axios from "axios";
`
				return os.WriteFile(filepath.Join(dir, "index.js"), []byte(srcContent), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name: "surface with path",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"react": "^18.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				srcContent := `import React from "react";`
				return os.WriteFile(filepath.Join(dir, "app.js"), []byte(srcContent), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name:        "surface non-existent path",
			path:        "/non/existent/path",
			expectError: true,
		},
		{
			name: "surface with config",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21", "axios": "^1.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				srcContent := `import { debounce } from "lodash";`
				if err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(srcContent), 0644); err != nil {
					return err
				}
				configContent := `ignore:
  - axios
`
				return os.WriteFile(filepath.Join(dir, ".depx.yml"), []byte(configContent), 0644)
			},
			path:        ".",
			configPath:  ".depx.yml",
			expectError: false,
		},
		{
			name: "surface with indirect flag",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				srcContent := `import { debounce } from "lodash";`
				if err := os.WriteFile(filepath.Join(dir, "index.js"), []byte(srcContent), 0644); err != nil {
					return err
				}
				// 创建 package-lock.json
				lockContent := `{
	"name": "test",
	"version": "1.0.0",
	"lockfileVersion": 2,
	"packages": {
		"": {
			"dependencies": {
				"lodash": "^4.17.21"
			}
		},
		"node_modules/lodash": {
			"version": "4.17.21"
		},
		"node_modules/left-pad": {
			"version": "1.3.0"
		}
	}
}`
				return os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte(lockContent), 0644)
			},
			path:         ".",
			showIndirect: true,
			expectError:  false,
		},
		{
			name: "surface with types package",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"@types/node": "^18.0.0", "lodash": "^4.17.21"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				srcContent := `import { debounce } from "lodash";`
				return os.WriteFile(filepath.Join(dir, "index.js"), []byte(srcContent), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name:        "surface with invalid config",
			setup:       func(dir string) error { return nil },
			path:        ".",
			configPath:  "/non/existent/config.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			defer os.Chdir(origDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// 捕获 stdout（同时重定向 color 输出）
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			oldColorOut := color.Output
			color.Output = w

			err := runSurface(tt.path, tt.configPath, tt.showIndirect, false)

			// 恢复 stdout 并读取内容
			_ = w.Close()
			color.Output = oldColorOut
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			os.Stdout = oldStdout
			out := buf.String()
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				// 对非错误情况断言输出包含 Runtime Surface 和 Most Critical
				if !tt.showIndirect && !tt.expectError {
					if !strings.Contains(out, "Most Critical") {
						t.Errorf("expected output to contain 'Most Critical', got: %s", out)
					}
					if !strings.Contains(out, "Runtime Surface") {
						t.Errorf("expected output to contain 'Runtime Surface', got: %s", out)
					}
				}
			}
		})
	}
}

func TestSurfaceCommandExecute(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	srcContent := `import { debounce } from "lodash";`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	surfaceCmd.ResetFlags()
	surfaceCmd.Flags().StringVarP(&surfaceConfigPath, "config", "c", "", "配置文件路径 (.depx.yml)")
	surfaceCmd.Flags().BoolVarP(&surfaceIndirect, "indirect", "i", false, "显示共享间接依赖摘要")
	surfaceCmd.Flags().BoolVarP(&surfaceDev, "dev", "D", false, "包含 devDependencies（默认仅分析 runtime dependencies）")

	// capture stdout and stderr (同时重定向 color 输出)
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	oldColorOut := color.Output
	color.Output = wOut

	if err := runSurface(".", "", false, false); err != nil {
		t.Fatalf("runSurface error: %v", err)
	}

	_ = wOut.Close()
	_ = wErr.Close()
	color.Output = oldColorOut
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	_, _ = io.Copy(&buf, rErr)
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	out := buf.String()

	if !strings.Contains(out, "Most Critical") {
		t.Fatalf("expected output to contain 'Most Critical', got: %s", out)
	}

	if !strings.Contains(out, "Runtime Surface") {
		t.Fatalf("expected output to contain 'Runtime Surface', got: %s", out)
	}
}

func TestSurfaceCommandWithIndirectFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	srcContent := `import { debounce } from "lodash";`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	lockContent := `{
	"name": "test",
	"version": "1.0.0",
	"lockfileVersion": 2,
	"packages": {
		"": {
			"dependencies": {
				"lodash": "^4.17.21"
			}
		},
		"node_modules/lodash": {
			"version": "4.17.21"
		},
		"node_modules/left-pad": {
			"version": "1.3.0"
		}
	}
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(lockContent), 0644); err != nil {
		t.Fatal(err)
	}

	surfaceCmd.ResetFlags()
	surfaceCmd.Flags().StringVarP(&surfaceConfigPath, "config", "c", "", "配置文件路径 (.depx.yml)")
	surfaceCmd.Flags().BoolVarP(&surfaceIndirect, "indirect", "i", false, "显示共享间接依赖摘要")
	surfaceCmd.Flags().BoolVarP(&surfaceDev, "dev", "D", false, "包含 devDependencies（默认仅分析 runtime dependencies）")

	// capture stdout and stderr (同时重定向 color 输出)
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	oldColorOut := color.Output
	color.Output = wOut

	if err := runSurface(".", "", true, false); err != nil {
		t.Fatalf("runSurface error: %v", err)
	}

	_ = wOut.Close()
	_ = wErr.Close()
	color.Output = oldColorOut
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, rOut)
	_, _ = io.Copy(&buf, rErr)
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	out := buf.String()

	if !strings.Contains(out, "Indirect Packages") {
		t.Fatalf("expected output to contain 'Indirect Packages', got: %s", out)
	}
	if !strings.Contains(out, "Total:") {
		t.Fatalf("expected output to contain indirect total, got: %s", out)
	}
	if strings.Contains(out, "left-pad") {
		t.Fatalf("expected indirect list to omit per-package surface details, got: %s", out)
	}
}

func TestSurfaceExcludesDevDependenciesByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pkgContent := `{
		"name": "test",
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"eslint": "^8.0.0", "typescript": "^5.0.0"}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	srcContent := `import React from "react";`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	oldColorOut := color.Output
	color.Output = w

	if err := runSurface(".", "", false, false); err != nil {
		t.Fatalf("runSurface error: %v", err)
	}

	_ = w.Close()
	color.Output = oldColorOut
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	out := buf.String()

	if !strings.Contains(out, "react") {
		t.Fatalf("expected runtime dependency react in output, got: %s", out)
	}
	for _, devPkg := range []string{"eslint", "typescript"} {
		if strings.Contains(out, devPkg) {
			t.Fatalf("expected dev dependency %q to be excluded by default, got: %s", devPkg, out)
		}
	}
}

func TestSurfaceIncludesDevDependenciesWithFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	pkgContent := `{
		"name": "test",
		"dependencies": {"react": "^18.0.0"},
		"devDependencies": {"eslint": "^8.0.0", "typescript": "^5.0.0"}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	srcContent := `import React from "react";
import eslint from "eslint";
import ts from "typescript";
`
	if err := os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte(srcContent), 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	oldColorOut := color.Output
	color.Output = w

	if err := runSurface(".", "", false, true); err != nil {
		t.Fatalf("runSurface error: %v", err)
	}

	_ = w.Close()
	color.Output = oldColorOut
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	out := buf.String()

	for _, pkg := range []string{"react", "eslint", "typescript"} {
		if !strings.Contains(out, pkg) {
			t.Fatalf("expected dependency %q in --dev output, got: %s", pkg, out)
		}
	}
}
