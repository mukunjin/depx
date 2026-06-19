package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonAnalyzer(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建 Python 源文件
	code := `import requests
from flask import Flask
import numpy as np

def main():
    print("Hello")
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewPythonAnalyzer()
	deps := []string{"requests", "flask", "numpy", "unused-package"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 检查使用的依赖
	if !result["requests"].Used {
		t.Error("requests should be marked as used")
	}
	if !result["flask"].Used {
		t.Error("flask should be marked as used")
	}
	if !result["numpy"].Used {
		t.Error("numpy should be marked as used")
	}

	// 检查未使用的依赖
	if result["unused-package"].Used {
		t.Error("unused-package should not be marked as used")
	}
}

func TestPythonAnalyzerSubpath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	code := `from requests.auth import HTTPBasicAuth
import numpy.linalg as la
`
	if err := os.WriteFile(filepath.Join(tmpDir, "lib.py"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewPythonAnalyzer()
	deps := []string{"requests", "numpy"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !result["requests"].Used {
		t.Error("requests should be marked as used")
	}
	if !result["numpy"].Used {
		t.Error("numpy should be marked as used")
	}
}

func TestPythonAnalyzerSkipDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建 venv 目录（应该被跳过）
	venvDir := filepath.Join(tmpDir, "venv")
	if err := os.MkdirAll(venvDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 在 venv 目录中创建文件
	venvCode := `import should_be_ignored`
	if err := os.WriteFile(filepath.Join(venvDir, "ignored.py"), []byte(venvCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 在根目录创建文件
	rootCode := `import actual_dep`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte(rootCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewPythonAnalyzer()
	deps := []string{"actual_dep", "should_be_ignored"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !result["actual_dep"].Used {
		t.Error("actual_dep should be marked as used")
	}
	if result["should_be_ignored"].Used {
		t.Error("should_be_ignored should not be marked as used (venv dir should be skipped)")
	}
}

func TestExtractPythonImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "Simple import",
			code:     `import requests`,
			expected: []string{"requests"},
		},
		{
			name:     "From import",
			code:     `from flask import Flask`,
			expected: []string{"flask"},
		},
		{
			name:     "Multiple imports",
			code:     "import requests\nfrom flask import Flask\nimport numpy",
			expected: []string{"requests", "flask", "numpy"},
		},
		{
			name:     "Aliased import",
			code:     `import numpy as np`,
			expected: []string{"numpy"},
		},
		{
			name:     "With comments",
			code:     "# This is a comment\nimport requests",
			expected: []string{"requests"},
		},
		{
			name:     "String literal",
			code:     "s = \"import fake\"\nimport real",
			expected: []string{"real"},
		},
		{
			name:     "Empty code",
			code:     ``,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractPythonImports(tt.code)

			if tt.expected == nil {
				if len(imports) != 0 {
					t.Errorf("Expected 0 imports, got %d: %v", len(imports), imports)
				}
				return
			}

			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(imports), imports)
				return
			}

			for i, exp := range tt.expected {
				if imports[i] != exp {
					t.Errorf("Expected import[%d] = %s, got %s", i, exp, imports[i])
				}
			}
		})
	}
}

func TestResolvePythonPackageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		importPath string
		expected   string
	}{
		{"Simple", "requests", "requests"},
		{"With submodule", "requests.auth", "requests"},
		{"Deep path", "numpy.linalg.svd", "numpy"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePythonPackageName(tt.importPath)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRemovePythonComments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comments",
			input:    `import requests`,
			expected: `import requests`,
		},
		{
			name:     "Line comment",
			input:    "# comment\nimport requests",
			expected: "\nimport requests",
		},
		{
			name:     "Comment in string",
			input:    `s = "import fake"`,
			expected: `s = "import fake"`,
		},
		{
			name:     "Empty",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removePythonComments(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
