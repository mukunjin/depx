package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNpmManifest(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		checkDeps   func(deps []string) error
	}{
		{
			name: "standard package.json",
			content: `{
				"name": "test-project",
				"version": "1.0.0",
				"dependencies": {
					"axios": "^1.0.0",
					"lodash": "^4.17.21"
				},
				"devDependencies": {
					"jest": "^29.0.0"
				}
			}`,
			expectError: false,
			checkDeps: func(deps []string) error {
				if len(deps) != 2 {
					return &testError{msg: "expected 2 runtime deps"}
				}
				depMap := toMap(deps)
				for _, pkg := range []string{"axios", "lodash"} {
					if !depMap[pkg] {
						return &testError{msg: "missing " + pkg}
					}
				}
				return nil
			},
		},
		{
			name:        "empty dependencies",
			content:     `{"name": "empty", "version": "1.0.0"}`,
			expectError: false,
			checkDeps: func(deps []string) error {
				if len(deps) != 0 {
					return &testError{msg: "expected 0 deps"}
				}
				return nil
			},
		},
		{
			name: "only devDependencies",
			content: `{
				"name": "dev-only",
				"devDependencies": {
					"jest": "^29.0.0",
					"eslint": "^8.0.0"
				}
			}`,
			expectError: false,
			checkDeps: func(deps []string) error {
				if len(deps) != 0 {
					return &testError{msg: "expected 0 runtime deps"}
				}
				return nil
			},
		},
		{
			name: "scoped packages",
			content: `{
				"name": "scoped",
				"dependencies": {
					"@types/node": "^18.0.0",
					"@babel/core": "^7.0.0"
				}
			}`,
			expectError: false,
			checkDeps: func(deps []string) error {
				if len(deps) != 2 {
					return &testError{msg: "expected 2 deps"}
				}
				depMap := toMap(deps)
				if !depMap["@types/node"] || !depMap["@babel/core"] {
					return &testError{msg: "missing scoped packages"}
				}
				return nil
			},
		},
		{
			name:        "invalid JSON",
			content:     `{invalid json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pkgPath := filepath.Join(tmpDir, "package.json")
			if err := os.WriteFile(pkgPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			manifest, err := NewNpmManifest(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("NewNpmManifest failed: %v", err)
				}
				return
			}

			if manifest.Type() != "npm" {
				t.Errorf("Expected type 'npm', got '%s'", manifest.Type())
			}

			deps, err := manifest.Dependencies()
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Dependencies failed: %v", err)
				}
				return
			}

			if tt.checkDeps != nil {
				if err := tt.checkDeps(deps); err != nil {
					t.Error(err)
				}
			}

			if tt.name == "standard package.json" {
				devDeps, err := manifest.DevDependencies()
				if err != nil {
					t.Fatalf("DevDependencies failed: %v", err)
				}
				if len(devDeps) != 1 || devDeps[0] != "jest" {
					t.Errorf("expected dev dependency jest, got %v", devDeps)
				}
			}

			if tt.name == "only devDependencies" {
				devDeps, err := manifest.DevDependencies()
				if err != nil {
					t.Fatalf("DevDependencies failed: %v", err)
				}
				if len(devDeps) != 2 {
					t.Errorf("expected 2 dev deps, got %d", len(devDeps))
				}
			}
		})
	}
}

func TestNpmManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewNpmManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing package.json")
	}
}

func TestGoModManifest(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		expectedLen int
		checkDeps   func(deps []string) error
	}{
		{
			name: "standard go.mod",
			content: `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
	golang.org/x/text v0.14.0 // indirect
)`,
			expectError: false,
			expectedLen: 2,
			checkDeps: func(deps []string) error {
				depMap := toMap(deps)
				if !depMap["github.com/gin-gonic/gin"] {
					return &testError{msg: "missing gin"}
				}
				if !depMap["github.com/spf13/cobra"] {
					return &testError{msg: "missing cobra"}
				}
				if depMap["golang.org/x/text"] {
					return &testError{msg: "indirect dep should be excluded"}
				}
				return nil
			},
		},
		{
			name: "empty go.mod",
			content: `module example.com/test

go 1.21`,
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "only indirect deps",
			content: `module example.com/test

go 1.21

require (
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)`,
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "single require",
			content: `module example.com/test

go 1.21

require github.com/gin-gonic/gin v1.9.1`,
			expectError: false,
			expectedLen: 1,
			checkDeps: func(deps []string) error {
				if len(deps) != 1 || deps[0] != "github.com/gin-gonic/gin" {
					return &testError{msg: "wrong dep"}
				}
				return nil
			},
		},
		{
			name: "mixed direct and indirect",
			content: `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	golang.org/x/text v0.14.0 // indirect
	github.com/spf13/cobra v1.8.0
)`,
			expectError: false,
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			modPath := filepath.Join(tmpDir, "go.mod")
			if err := os.WriteFile(modPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			manifest, err := NewGoModManifest(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("NewGoModManifest failed: %v", err)
				}
				return
			}

			if manifest.Type() != "go" {
				t.Errorf("Expected type 'go', got '%s'", manifest.Type())
			}

			deps, err := manifest.Dependencies()
			if err != nil {
				if !tt.expectError {
					t.Fatalf("Dependencies failed: %v", err)
				}
				return
			}

			if len(deps) != tt.expectedLen {
				t.Errorf("Expected %d deps, got %d: %v", tt.expectedLen, len(deps), deps)
			}

			if tt.checkDeps != nil {
				if err := tt.checkDeps(deps); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGoModManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := NewGoModManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing go.mod")
	}
}

func TestNpmManifest_DevDependenciesEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		checkDev    func(deps []string) error
	}{
		{
			name: "empty devDependencies",
			content: `{
				"name": "test",
				"devDependencies": {}
			}`,
			expectError: false,
			checkDev: func(deps []string) error {
				if len(deps) != 0 {
					return &testError{msg: "expected 0 dev deps"}
				}
				return nil
			},
		},
		{
			name: "no devDependencies field",
			content: `{
				"name": "test",
				"dependencies": {"react": "^18.0.0"}
			}`,
			expectError: false,
			checkDev: func(deps []string) error {
				if len(deps) != 0 {
					return &testError{msg: "expected 0 dev deps when field missing"}
				}
				return nil
			},
		},
		{
			name: "multiple devDependencies",
			content: `{
				"name": "test",
				"devDependencies": {
					"jest": "^29.0.0",
					"eslint": "^8.0.0",
					"prettier": "^3.0.0"
				}
			}`,
			expectError: false,
			checkDev: func(deps []string) error {
				if len(deps) != 3 {
					return &testError{msg: "expected 3 dev deps"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pkgPath := filepath.Join(tmpDir, "package.json")
			if err := os.WriteFile(pkgPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			manifest, err := NewNpmManifest(tmpDir)
			if err != nil {
				if !tt.expectError {
					t.Fatalf("NewNpmManifest failed: %v", err)
				}
				return
			}

			devDeps, err := manifest.DevDependencies()
			if err != nil {
				if !tt.expectError {
					t.Fatalf("DevDependencies failed: %v", err)
				}
				return
			}

			if tt.checkDev != nil {
				if err := tt.checkDev(devDeps); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestGoModManifest_MultipleRequireBlocks(t *testing.T) {
	tmpDir := t.TempDir()
	content := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)

require (
	github.com/spf13/cobra v1.8.0
	golang.org/x/text v0.14.0 // indirect
)
`
	modPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewGoModManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewGoModManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 应该包含 gin 和 cobra，排除 indirect 的 x/text
	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d: %v", len(deps), deps)
	}

	depMap := toMap(deps)
	if !depMap["github.com/gin-gonic/gin"] {
		t.Error("missing gin")
	}
	if !depMap["github.com/spf13/cobra"] {
		t.Error("missing cobra")
	}
	if depMap["golang.org/x/text"] {
		t.Error("indirect dep should be excluded")
	}
}

func TestGoModManifest_RequireWithoutParentheses(t *testing.T) {
	tmpDir := t.TempDir()
	content := `module example.com/test

go 1.21

require github.com/gin-gonic/gin v1.9.1
require github.com/spf13/cobra v1.8.0
`
	modPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewGoModManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewGoModManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d: %v", len(deps), deps)
	}
}

func TestCargoManifest_WorkspaceDependencies(t *testing.T) {
	tmpDir := t.TempDir()
	cargoToml := `[package]
name = "test"
version = "0.1.0"

[dependencies]
serde = { version = "1.0", features = ["derive"] }
tokio = { version = "1.0", default-features = false }
reqwest = "0.11"

[dev-dependencies]
mockall = "0.11"
`
	cargoPath := filepath.Join(tmpDir, "Cargo.toml")
	if err := os.WriteFile(cargoPath, []byte(cargoToml), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewCargoManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewCargoManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 3 {
		t.Errorf("Expected 3 deps, got %d: %v", len(deps), deps)
	}
}

func TestPipManifest_CommentsAndEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	reqTxt := `# This is a comment
requests==2.28.0

# Another comment
numpy>=1.21.0

`
	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte(reqTxt), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewPipManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewPipManifest failed: %v", err)
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 2 {
		t.Errorf("Expected 2 deps, got %d: %v", len(deps), deps)
	}
}

func TestPipManifest_VersionConstraints(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"exact version", "requests==2.28.0", "requests"},
		{"minimum version", "numpy>=1.21.0", "numpy"},
		{"maximum version", "pandas<=1.3.0", "pandas"},
		{"compatible release", "flask~=2.0.0", "flask"},
		{"not equal", "django!=3.0.0", "django"},
		{"greater than", "package>1.0", "package"},
		{"less than", "package<2.0", "package"},
		{"no version", "requests", "requests"},
		{"with extras", "requests[security]==2.28.0", "requests"},
		{"with spaces", " requests >= 2.28.0 ", "requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPipPackageName(tt.line)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Helper types and functions
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func toMap(slice []string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range slice {
		m[s] = true
	}
	return m
}
