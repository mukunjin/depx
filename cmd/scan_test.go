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

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err := runScan(tt.path, tt.configPath)
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

	// 设置参数并执行
	scanCmd.SetArgs([]string{"."})
	if err := scanCmd.Execute(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
