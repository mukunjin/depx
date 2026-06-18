package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	if len(cfg.ExcludeDirs) == 0 {
		t.Error("Default config should have exclude dirs")
	}
	if !cfg.LockFile {
		t.Error("LockFile should be true by default")
	}
	if cfg.ReadNodeModules {
		t.Error("ReadNodeModules should be false by default")
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		check       func(*Config) error
	}{
		{
			name: "full config",
			content: `
ignore:
  - "@types/node"
  - "typescript"
exclude_dirs:
  - "custom_dir"
exclude_files:
  - "*.test.js"
read_node_modules: true
lock_file: false`,
			expectError: false,
			check: func(cfg *Config) error {
				if len(cfg.Ignore) != 2 || cfg.Ignore[0] != "@types/node" {
					return errTest("ignore mismatch")
				}
				if len(cfg.ExcludeDirs) != 1 || cfg.ExcludeDirs[0] != "custom_dir" {
					return errTest("exclude_dirs mismatch")
				}
				if !cfg.ReadNodeModules {
					return errTest("read_node_modules should be true")
				}
				if cfg.LockFile {
					return errTest("lock_file should be false")
				}
				return nil
			},
		},
		{
			name:        "empty config",
			content:     "",
			expectError: false,
			check: func(cfg *Config) error {
				if len(cfg.Ignore) != 0 {
					return errTest("expected empty ignore")
				}
				return nil
			},
		},
		{
			name:        "invalid YAML",
			content:     "invalid: yaml: content:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".depx.yml")
			if err := os.WriteFile(configPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			cfg, err := Load(configPath)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Load failed: %v", err)
				}
				return
			}

			if tt.check != nil {
				if err := tt.check(cfg); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestLoadConfigMissing(t *testing.T) {
	_, err := Load("/nonexistent/path/.depx.yml")
	if err == nil {
		t.Error("Expected error for missing config file")
	}
}

func TestFindAndLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// 测试配置文件不存在的情况
	cfg, err := FindAndLoad(tmpDir)
	if err != nil {
		t.Fatalf("FindAndLoad failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("FindAndLoad returned nil")
	}
	if !cfg.LockFile {
		t.Error("Default config should have LockFile = true")
	}

	// 创建配置文件
	configContent := `ignore: ["test-pkg"]`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 测试配置文件存在的情况
	cfg, err = FindAndLoad(tmpDir)
	if err != nil {
		t.Fatalf("FindAndLoad failed: %v", err)
	}
	if len(cfg.Ignore) != 1 || cfg.Ignore[0] != "test-pkg" {
		t.Errorf("Expected ignore=['test-pkg'], got %v", cfg.Ignore)
	}
}

func TestIsIgnored(t *testing.T) {
	cfg := &Config{Ignore: []string{"@types/node", "typescript"}}

	tests := []struct {
		pkg      string
		expected bool
	}{
		{"@types/node", true},
		{"typescript", true},
		{"react", false},
		{"lodash", false},
	}

	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			if cfg.IsIgnored(tt.pkg) != tt.expected {
				t.Errorf("IsIgnored(%s) = %v, expected %v", tt.pkg, !tt.expected, tt.expected)
			}
		})
	}
}

func TestIsDirExcluded(t *testing.T) {
	cfg := &Config{ExcludeDirs: []string{"node_modules", "vendor"}}

	tests := []struct {
		dir      string
		expected bool
	}{
		{"node_modules", true},
		{"vendor", true},
		{"src", false},
		{"dist", false},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			if cfg.IsDirExcluded(tt.dir) != tt.expected {
				t.Errorf("IsDirExcluded(%s) = %v, expected %v", tt.dir, !tt.expected, tt.expected)
			}
		})
	}
}

type testError string

func (e testError) Error() string { return string(e) }

func errTest(msg string) error { return testError(msg) }
