package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMergeWithDev(t *testing.T) {
	tmpDir := t.TempDir()
	content := `{
		"name": "test",
		"dependencies": {"react": "^18.0.0", "lodash": "^4.0.0"},
		"devDependencies": {"eslint": "^8.0.0", "lodash": "^4.0.0"}
	}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := NewNpmManifest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	deps, err := MergeWithDev(m)
	if err != nil {
		t.Fatal(err)
	}

	if len(deps) != 3 {
		t.Fatalf("expected 3 merged deps, got %d: %v", len(deps), deps)
	}

	seen := toMap(deps)
	for _, pkg := range []string{"react", "lodash", "eslint"} {
		if !seen[pkg] {
			t.Errorf("missing merged dependency %q", pkg)
		}
	}
}

func TestNpmManifestSingleRead(t *testing.T) {
	tmpDir := t.TempDir()
	content := `{"dependencies": {"axios": "^1.0.0"}, "devDependencies": {"jest": "^29.0.0"}}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := NewNpmManifest(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := m.Dependencies(); err != nil {
		t.Fatal(err)
	}
	if _, err := m.DevDependencies(); err != nil {
		t.Fatal(err)
	}
	if m.pkg == nil {
		t.Fatal("expected package.json to be cached after first read")
	}
}
