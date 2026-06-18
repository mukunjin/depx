package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPipManifest(t *testing.T) {
	tmpDir := t.TempDir()

	reqTxt := `requests==2.28.0
numpy>=1.21.0
pandas
flask~=2.0.0
# This is a comment
django!=3.0.0
`

	reqPath := filepath.Join(tmpDir, "requirements.txt")
	if err := os.WriteFile(reqPath, []byte(reqTxt), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewPipManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewPipManifest failed: %v", err)
	}

	if manifest.Type() != "pip" {
		t.Errorf("Expected type 'pip', got '%s'", manifest.Type())
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	if len(deps) != 5 {
		t.Errorf("Expected 5 dependencies, got %d: %v", len(deps), deps)
	}

	expected := map[string]bool{"requests": true, "numpy": true, "pandas": true, "flask": true, "django": true}
	for _, dep := range deps {
		if !expected[dep] {
			t.Errorf("Unexpected dependency: %s", dep)
		}
	}
}

func TestPipManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := NewPipManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing requirements.txt")
	}
}

func TestPipManifestEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	reqTxt := `# Only comments
# No dependencies
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

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestPipManifestWithOptions(t *testing.T) {
	tmpDir := t.TempDir()

	reqTxt := `requests==2.28.0
-r other-requirements.txt
-e git+https://github.com/user/repo.git
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

	// 应该只提取 requests 和 numpy，忽略 -r 和 -e
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d: %v", len(deps), deps)
	}
}

func TestExtractPipPackageName(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"Exact version", "requests==2.28.0", "requests"},
		{"Greater equal", "numpy>=1.21.0", "numpy"},
		{"Less equal", "pandas<=1.3.0", "pandas"},
		{"Compatible", "flask~=2.0.0", "flask"},
		{"Not equal", "django!=3.0.0", "django"},
		{"Greater", "package>1.0", "package"},
		{"Less", "package<2.0", "package"},
		{"No version", "requests", "requests"},
		{"With spaces", " requests == 2.28.0 ", "requests"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPipPackageName(tt.line)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
