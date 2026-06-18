package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNpmManifest(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 测试用例 1: 标准 package.json
	pkgJSON := `{
		"name": "test-project",
		"version": "1.0.0",
		"dependencies": {
			"axios": "^1.0.0",
			"lodash": "^4.17.21"
		},
		"devDependencies": {
			"jest": "^29.0.0"
		}
	}`

	pkgPath := filepath.Join(tmpDir, "package.json")
	if err := os.WriteFile(pkgPath, []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewNpmManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewNpmManifest failed: %v", err)
	}

	if manifest.Type() != "npm" {
		t.Errorf("Expected type 'npm', got '%s'", manifest.Type())
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 应该包含 3 个依赖
	if len(deps) != 3 {
		t.Errorf("Expected 3 dependencies, got %d", len(deps))
	}

	// 检查依赖是否包含
	depMap := make(map[string]bool)
	for _, dep := range deps {
		depMap[dep] = true
	}

	expected := []string{"axios", "lodash", "jest"}
	for _, exp := range expected {
		if !depMap[exp] {
			t.Errorf("Expected dependency '%s' not found", exp)
		}
	}
}

func TestNpmManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := NewNpmManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing package.json, got nil")
	}
}

func TestGoModManifest(t *testing.T) {
	tmpDir := t.TempDir()

	goMod := `module example.com/test

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/spf13/cobra v1.8.0
	golang.org/x/text v0.14.0 // indirect
)
`

	modPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(modPath, []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := NewGoModManifest(tmpDir)
	if err != nil {
		t.Fatalf("NewGoModManifest failed: %v", err)
	}

	if manifest.Type() != "go" {
		t.Errorf("Expected type 'go', got '%s'", manifest.Type())
	}

	deps, err := manifest.Dependencies()
	if err != nil {
		t.Fatalf("Dependencies failed: %v", err)
	}

	// 应该包含 2 个直接依赖（排除 indirect）
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d: %v", len(deps), deps)
	}

	depMap := make(map[string]bool)
	for _, dep := range deps {
		depMap[dep] = true
	}

	if !depMap["github.com/gin-gonic/gin"] {
		t.Error("Expected 'github.com/gin-gonic/gin' not found")
	}
	if !depMap["github.com/spf13/cobra"] {
		t.Error("Expected 'github.com/spf13/cobra' not found")
	}
	if depMap["golang.org/x/text"] {
		t.Error("Indirect dependency 'golang.org/x/text' should not be included")
	}
}

func TestGoModManifestMissing(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := NewGoModManifest(tmpDir)
	if err == nil {
		t.Error("Expected error for missing go.mod, got nil")
	}
}
