package lockfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNpmLockFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedLen int
		check       func(deps []Dependency) error
	}{
		{
			name: "lockfileVersion 2",
			content: `{
  "name": "test-project",
  "version": "1.0.0",
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "test-project", "version": "1.0.0"},
    "node_modules/lodash": {
      "version": "4.17.21",
      "resolved": "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz"
    },
    "node_modules/express": {
      "version": "4.18.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz"
    }
  }
}`,
			expectError: false,
			expectedLen: 2,
			check: func(deps []Dependency) error {
				found := toMap(deps)
				if !found["lodash"] {
					return errTest("missing lodash")
				}
				if !found["express"] {
					return errTest("missing express")
				}
				return nil
			},
		},
		{
			name: "lockfileVersion 1",
			content: `{
  "name": "test-project",
  "version": "1.0.0",
  "lockfileVersion": 1,
  "dependencies": {
    "lodash": {
      "version": "4.17.21",
      "resolved": "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz"
    },
    "express": {
      "version": "4.18.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.18.2.tgz"
    }
  }
}`,
			expectError: false,
			expectedLen: 2,
		},
		{
			name: "nested node_modules paths",
			content: `{
  "name": "test",
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "test"},
    "node_modules/foo": {"version": "1.0.0"},
    "node_modules/bar": {"version": "1.0.0"}
  }
}`,
			expectError: false,
			expectedLen: 2,
			check: func(deps []Dependency) error {
				found := toMap(deps)
				if !found["foo"] || !found["bar"] {
					return errTest("missing foo or bar")
				}
				return nil
			},
		},
		{
			name: "scoped packages",
			content: `{
  "lockfileVersion": 2,
  "packages": {
    "": {},
    "node_modules/@babel/core": {"version": "7.0.0"},
    "node_modules/@types/node": {"version": "18.0.0"}
  }
}`,
			expectError: false,
			expectedLen: 2,
			check: func(deps []Dependency) error {
				found := toMap(deps)
				if !found["@babel/core"] || !found["@types/node"] {
					return errTest("missing scoped packages")
				}
				return nil
			},
		},
		{
			name:        "invalid JSON",
			content:     `{invalid}`,
			expectError: true,
		},
		{
			name: "empty packages",
			content: `{
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "empty"}
  }
}`,
			expectError: false,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			lockPath := filepath.Join(tmpDir, "package-lock.json")
			if err := os.WriteFile(lockPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			lf, err := NewNpmLockFile(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("NewNpmLockFile failed: %v", err)
				}
				return
			}

			if lf.Type() != "npm" {
				t.Errorf("Expected type 'npm', got '%s'", lf.Type())
			}

			deps, err := lf.Dependencies()
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Dependencies failed: %v", err)
				}
				return
			}

			if len(deps) != tt.expectedLen {
				t.Errorf("Expected %d deps, got %d", tt.expectedLen, len(deps))
			}

			if tt.check != nil {
				if err := tt.check(deps); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGoLockFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedLen int
	}{
		{
			name: "standard go.sum",
			content: `github.com/gin-gonic/gin v1.9.1 h1:abc123=
github.com/gin-gonic/gin v1.9.1/go.mod h1:def456=
github.com/stretchr/testify v1.8.4 h1:ghi789=
github.com/stretchr/testify v1.8.4/go.mod h1:jkl012=
`,
			expectError: false,
			expectedLen: 2,
		},
		{
			name:        "empty go.sum",
			content:     "",
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "single package",
			content: `github.com/gin-gonic/gin v1.9.1 h1:abc=
github.com/gin-gonic/gin v1.9.1/go.mod h1:def=
`,
			expectError: false,
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			lf, err := NewGoLockFile(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("NewGoLockFile failed: %v", err)
				}
				return
			}

			if lf.Type() != "go" {
				t.Errorf("Expected type 'go', got '%s'", lf.Type())
			}

			deps, err := lf.Dependencies()
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Dependencies failed: %v", err)
				}
				return
			}

			if len(deps) != tt.expectedLen {
				t.Errorf("Expected %d deps, got %d", tt.expectedLen, len(deps))
			}
		})
	}
}

func TestGoLockFileBasic(t *testing.T) {
	sumContent := `github.com/gin-gonic/gin v1.9.1 h1:abc123=
github.com/gin-gonic/gin v1.9.1/go.mod h1:def456=
github.com/stretchr/testify v1.8.4 h1:ghi789=
github.com/stretchr/testify v1.8.4/go.mod h1:jkl012=
`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte(sumContent), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewGoLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewGoLockFile failed: %v", err)
	}

	if lf.Type() != "go" {
		t.Errorf("Expected type 'go', got '%s'", lf.Type())
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d", len(deps))
	}

	found := toMap(deps)
	if !found["github.com/gin-gonic/gin"] {
		t.Error("missing gin")
	}
	if !found["github.com/stretchr/testify"] {
		t.Error("missing testify")
	}
}

func TestRustLockFile(t *testing.T) {
	lockContent := `# This file is automatically @generated by Cargo.
version = 3

[[package]]
name = "serde"
version = "1.0.193"

[[package]]
name = "tokio"
version = "1.35.0"

[[package]]
name = "my-project"
version = "0.1.0"
`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.lock"), []byte(lockContent), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewRustLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewRustLockFile failed: %v", err)
	}

	if lf.Type() != "cargo" {
		t.Errorf("Expected type 'cargo', got '%s'", lf.Type())
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 3 {
		t.Errorf("Expected 3 deps, got %d", len(deps))
	}

	found := toMap(deps)
	if !found["serde"] {
		t.Error("missing serde")
	}
	if !found["tokio"] {
		t.Error("missing tokio")
	}
}

func TestNpmLockFileSharedIndirectCounts(t *testing.T) {
	content := `{
  "name": "test",
  "version": "1.0.0",
  "lockfileVersion": 2,
  "packages": {
    "": {
      "dependencies": {
        "react": "^18.0.0",
        "react-dom": "^18.0.0"
      }
    },
    "node_modules/react": {
      "version": "18.0.0",
      "dependencies": {
        "loose-envify": "^1.1.0",
        "scheduler": "^0.23.0"
      }
    },
    "node_modules/react-dom": {
      "version": "18.0.0",
      "dependencies": {
        "loose-envify": "^1.1.0",
        "scheduler": "^0.23.0"
      }
    },
    "node_modules/loose-envify": {"version": "1.4.0"},
    "node_modules/scheduler": {"version": "0.23.0"}
  }
}`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewNpmLockFile(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	counts := lf.SharedIndirectCounts([]string{"react", "react-dom"})
	if counts["scheduler"] != 2 {
		t.Errorf("expected scheduler required by 2 direct packages, got %d", counts["scheduler"])
	}
	if counts["loose-envify"] != 2 {
		t.Errorf("expected loose-envify required by 2 direct packages, got %d", counts["loose-envify"])
	}
}

func TestDetectLockFile(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		expectType  string
		expectError bool
	}{
		{
			name: "npm lockfile",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package-lock.json"), []byte("{}"), 0644)
			},
			expectType:  "npm",
			expectError: false,
		},
		{
			name: "go lockfile",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "go.sum"), []byte(""), 0644)
			},
			expectType:  "go",
			expectError: false,
		},
		{
			name:        "no lockfile",
			setup:       func(dir string) error { return nil },
			expectType:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			if err := tt.setup(tmpDir); err != nil {
				t.Fatal(err)
			}

			lf, err := DetectLockFile(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("DetectLockFile failed: %v", err)
				}
				return
			}

			if lf.Type() != tt.expectType {
				t.Errorf("Expected type '%s', got '%s'", tt.expectType, lf.Type())
			}
		})
	}
}

func TestExtractPackagePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path",
			input:    "node_modules/lodash",
			expected: "lodash",
		},
		{
			name:     "scoped package",
			input:    "node_modules/@types/node",
			expected: "@types/node",
		},
		{
			name:     "nested node_modules",
			input:    "node_modules/foo/node_modules/bar",
			expected: "bar",
		},
		{
			name:     "no node_modules prefix",
			input:    "lodash",
			expected: "lodash",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPackagePath(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNpmLockFile_NestedNodeModules(t *testing.T) {
	content := `{
  "name": "test",
  "lockfileVersion": 2,
  "packages": {
    "": {"name": "test"},
    "node_modules/foo": {"version": "1.0.0"},
    "node_modules/foo/node_modules/bar": {"version": "2.0.0"},
    "node_modules/baz": {"version": "3.0.0"}
  }
}`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewNpmLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewNpmLockFile failed: %v", err)
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 应该去重，bar 只出现一次
	found := toMap(deps)
	if !found["foo"] || !found["bar"] || !found["baz"] {
		t.Errorf("Missing expected packages: %v", deps)
	}
}

func TestNpmLockFile_EmptyPackages(t *testing.T) {
	content := `{
  "name": "test",
  "lockfileVersion": 2,
  "packages": {}
}`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "package-lock.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewNpmLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewNpmLockFile failed: %v", err)
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 deps, got %d", len(deps))
	}
}

func TestGoLockFile_MalformedLines(t *testing.T) {
	content := `github.com/gin-gonic/gin v1.9.1 h1:abc=
malformed
github.com/stretchr/testify v1.8.4 h1:def=
`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.sum"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewGoLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewGoLockFile failed: %v", err)
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 应该跳过格式错误的行
	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d", len(deps))
	}
}

func TestRustLockFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.lock"), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewRustLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewRustLockFile failed: %v", err)
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 deps, got %d", len(deps))
	}
}

func TestRustLockFile_NoPackages(t *testing.T) {
	content := `# This file is automatically @generated by Cargo.
version = 3
`

	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.lock"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	lf, err := NewRustLockFile(tmpDir)
	if err != nil {
		t.Fatalf("NewRustLockFile failed: %v", err)
	}

	deps, err := lf.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 deps, got %d", len(deps))
	}
}

// Helper functions
type testError string

func (e testError) Error() string { return string(e) }

func errTest(msg string) error { return testError(msg) }

func toMap(deps []Dependency) map[string]bool {
	m := make(map[string]bool)
	for _, d := range deps {
		m[d.Name] = true
	}
	return m
}
