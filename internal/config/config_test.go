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
		{
			name: "config with only ignore",
			content: `
ignore:
  - "pkg1"
  - "pkg2"`,
			expectError: false,
			check: func(cfg *Config) error {
				if len(cfg.Ignore) != 2 {
					return errTest("ignore should have 2 items")
				}
				if !cfg.LockFile {
					return errTest("lock_file should default to true")
				}
				return nil
			},
		},
		{
			name: "config with only exclude_dirs",
			content: `
exclude_dirs:
  - "dir1"
  - "dir2"`,
			expectError: false,
			check: func(cfg *Config) error {
				if len(cfg.ExcludeDirs) != 2 {
					return errTest("exclude_dirs should have 2 items")
				}
				return nil
			},
		},
		{
			name: "config with only exclude_files",
			content: `
exclude_files:
  - "*.test.js"
  - "*.spec.ts"`,
			expectError: false,
			check: func(cfg *Config) error {
				if len(cfg.ExcludeFiles) != 2 {
					return errTest("exclude_files should have 2 items")
				}
				return nil
			},
		},
		{
			name: "config with read_node_modules true",
			content: `
read_node_modules: true`,
			expectError: false,
			check: func(cfg *Config) error {
				if !cfg.ReadNodeModules {
					return errTest("read_node_modules should be true")
				}
				return nil
			},
		},
		{
			name: "config with lock_file false",
			content: `
lock_file: false`,
			expectError: false,
			check: func(cfg *Config) error {
				if cfg.LockFile {
					return errTest("lock_file should be false")
				}
				return nil
			},
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

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config",
			config: &Config{
				Ignore:       []string{"pkg1"},
				ExcludeDirs:  []string{"node_modules"},
				ExcludeFiles: []string{"*.test.js"},
			},
			expectError: false,
		},
		{
			name: "empty exclude_dirs",
			config: &Config{
				ExcludeDirs: []string{""},
			},
			expectError: true,
		},
		{
			name: "empty exclude_files",
			config: &Config{
				ExcludeFiles: []string{""},
			},
			expectError: true,
		},
		{
			name: "multiple empty values",
			config: &Config{
				ExcludeDirs:  []string{"valid", ""},
				ExcludeFiles: []string{"*.js", ""},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidate_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "valid config with all fields",
			config: &Config{
				Ignore:          []string{"pkg1", "pkg2"},
				ExcludeDirs:     []string{"node_modules", "vendor"},
				ExcludeFiles:    []string{"*.test.js", "*.spec.ts"},
				ReadNodeModules: true,
				LockFile:        true,
			},
			expectError: false,
		},
		{
			name: "empty config",
			config: &Config{
				Ignore:       []string{},
				ExcludeDirs:  []string{},
				ExcludeFiles: []string{},
			},
			expectError: false,
		},
		{
			name: "exclude_dirs with empty string",
			config: &Config{
				ExcludeDirs: []string{""},
			},
			expectError: true,
		},
		{
			name: "exclude_files with empty string",
			config: &Config{
				ExcludeFiles: []string{""},
			},
			expectError: true,
		},
		{
			name: "exclude_dirs with multiple empty strings",
			config: &Config{
				ExcludeDirs: []string{"valid", "", "also-valid"},
			},
			expectError: true,
		},
		{
			name: "exclude_files with multiple empty strings",
			config: &Config{
				ExcludeFiles: []string{"*.js", "", "*.ts"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsIgnored_EmptyList(t *testing.T) {
	cfg := &Config{Ignore: []string{}}
	if cfg.IsIgnored("any-pkg") {
		t.Error("Should not ignore any package when ignore list is empty")
	}
}

func TestIsIgnored_CaseSensitive(t *testing.T) {
	cfg := &Config{Ignore: []string{"React"}}
	if cfg.IsIgnored("react") {
		t.Error("IsIgnored should be case-sensitive")
	}
	if !cfg.IsIgnored("React") {
		t.Error("Should ignore exact case match")
	}
}

func TestIsDirExcluded_EmptyList(t *testing.T) {
	cfg := &Config{ExcludeDirs: []string{}}
	if cfg.IsDirExcluded("node_modules") {
		t.Error("Should not exclude any dir when list is empty")
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// 验证默认值
	if len(cfg.ExcludeDirs) != 4 {
		t.Errorf("Expected 4 default exclude dirs, got %d", len(cfg.ExcludeDirs))
	}
	if !cfg.LockFile {
		t.Error("LockFile should default to true")
	}
	if cfg.ReadNodeModules {
		t.Error("ReadNodeModules should default to false")
	}
	if len(cfg.Ignore) != 0 {
		t.Errorf("Ignore should be empty by default, got %v", cfg.Ignore)
	}
	if len(cfg.ExcludeFiles) != 0 {
		t.Errorf("ExcludeFiles should be empty by default, got %v", cfg.ExcludeFiles)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".depx.yml")
	invalidContent := `invalid: yaml: content: [`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestFindAndLoad_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建配置文件
	configContent := `
ignore:
  - "@types/node"
  - "typescript"
exclude_dirs:
  - "custom_dir"
read_node_modules: true
lock_file: false
`
	configPath := filepath.Join(tmpDir, ".depx.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := FindAndLoad(tmpDir)
	if err != nil {
		t.Fatalf("FindAndLoad failed: %v", err)
	}

	if len(cfg.Ignore) != 2 {
		t.Errorf("Expected 2 ignore items, got %d", len(cfg.Ignore))
	}
	if len(cfg.ExcludeDirs) != 1 {
		t.Errorf("Expected 1 exclude_dir, got %d", len(cfg.ExcludeDirs))
	}
	if !cfg.ReadNodeModules {
		t.Error("ReadNodeModules should be true")
	}
	if cfg.LockFile {
		t.Error("LockFile should be false")
	}
}

type testError string

func (e testError) Error() string { return string(e) }

func errTest(msg string) error { return testError(msg) }
