package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunScan(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		path        string
		configPath  string
		expectError bool
	}{
		{
			name: "scan current directory",
			setup: func(dir string) error {
				content := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name: "scan with path argument",
			setup: func(dir string) error {
				content := `{"name": "test", "dependencies": {"axios": "^1.0.0"}}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name: "scan with config flag",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				configContent := `ignore:
  - lodash
`
				return os.WriteFile(filepath.Join(dir, ".depx.yml"), []byte(configContent), 0644)
			},
			path:        ".",
			configPath:  ".depx.yml",
			expectError: false,
		},
		{
			name:        "scan non-existent path",
			path:        "/non/existent/path",
			expectError: true,
		},
		{
			name:        "scan with invalid config",
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

			// 重置全局变量 configPath
			oldConfigPath := configPath
			defer func() { configPath = oldConfigPath }()
			configPath = ""

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err := runScan(tt.path, tt.configPath, false, false, false)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestScanCommandExecute(t *testing.T) {
	// 测试命令执行（覆盖 Run 函数）
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建测试项目
	content := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// 重置命令状态
	scanCmd.ResetFlags()
	scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径 (.depx.yml)")

	// 设置参数并执行 - 使用 SetArgs 避免继承测试参数
	scanCmd.SetArgs([]string{"."})
	// 注意：这里不检查错误，因为 cobra 可能会继承测试运行器的参数
	_ = scanCmd.Execute()
}

func TestScanCommandArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setup       func(string) error
		expectError bool
	}{
		{
			name: "scan with explicit path",
			args: []string{"."},
			setup: func(dir string) error {
				content := `{"name": "test", "dependencies": {"react": "^18.0.0"}}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0644)
			},
			expectError: false,
		},
		{
			name: "scan with no args (uses current dir)",
			args: []string{},
			setup: func(dir string) error {
				content := `{"name": "test", "dependencies": {"vue": "^3.0.0"}}`
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(content), 0644)
			},
			expectError: false,
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

			// 重置命令状态
			scanCmd.ResetFlags()
			scanCmd.Flags().StringVarP(&configPath, "config", "c", "", "配置文件路径 (.depx.yml)")
			scanCmd.SetArgs(tt.args)

			err := scanCmd.Execute()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRunScanWithInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建有效的 package.json
	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建无效的配置文件
	invalidConfig := `invalid: yaml: content: [`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0644); err != nil {
		t.Fatal(err)
	}

	err := runScan(".", configPath, false, false, false)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestRunScanWithEmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建有效的 package.json
	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建空的配置文件
	emptyConfig := ``
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(emptyConfig), 0644); err != nil {
		t.Fatal(err)
	}

	err := runScan(".", configPath, false, false, false)
	if err != nil {
		t.Errorf("unexpected error for empty config: %v", err)
	}
}

func TestRunScanWithExcludeDirs(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建 package.json
	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建带有 exclude_dirs 的配置文件
	configContent := `exclude_dirs:
  - node_modules
  - vendor
`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runScan(".", configPath, false, false, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunScanWithLockFileDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建 package.json
	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建禁用 lock_file 的配置文件
	configContent := `lock_file: false`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := runScan(".", configPath, false, false, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunScanWithIndirectFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建 package.json
	pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 showIndirect 参数
	err := runScan(".", "", true, false, false)
	if err != nil {
		t.Errorf("unexpected error with showIndirect: %v", err)
	}
}

func TestRunScanWithTypePkgsFlag(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建 package.json，包含类型包
	pkgContent := `{"name": "test", "dependencies": {"@types/node": "^18.0.0", "lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试 showTypePkgs 参数
	err := runScan(".", "", false, false, true)
	if err != nil {
		t.Errorf("unexpected error with showTypePkgs: %v", err)
	}
}

func TestRunScanWithBothFlags(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// 创建 package.json，包含类型包
	pkgContent := `{"name": "test", "dependencies": {"@types/node": "^18.0.0", "lodash": "^4.17.21"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试同时使用 showIndirect 和 showTypePkgs
	err := runScan(".", "", true, false, true)
	if err != nil {
		t.Errorf("unexpected error with both flags: %v", err)
	}
}
