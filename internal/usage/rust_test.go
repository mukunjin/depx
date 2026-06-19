package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRustAnalyzer(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建 Rust 源文件
	code := `use serde::Deserialize;
use tokio::sync::Mutex;
extern crate reqwest;

fn main() {
    println!("Hello");
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.rs"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewRustAnalyzer()
	deps := []string{"serde", "tokio", "reqwest", "unused-crate"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 检查使用的依赖
	if !result["serde"].Used {
		t.Error("serde should be marked as used")
	}
	if !result["tokio"].Used {
		t.Error("tokio should be marked as used")
	}
	if !result["reqwest"].Used {
		t.Error("reqwest should be marked as used")
	}

	// 检查未使用的依赖
	if result["unused-crate"].Used {
		t.Error("unused-crate should not be marked as used")
	}
}

func TestRustAnalyzerSubpath(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	code := `use serde_json::Value;
use tokio::sync::mpsc;
`
	if err := os.WriteFile(filepath.Join(tmpDir, "lib.rs"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewRustAnalyzer()
	deps := []string{"serde_json", "tokio"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !result["serde_json"].Used {
		t.Error("serde_json should be marked as used")
	}
	if !result["tokio"].Used {
		t.Error("tokio should be marked as used")
	}
}

func TestRustAnalyzerSkipDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// 创建 target 目录（应该被跳过）
	targetDir := filepath.Join(tmpDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 在 target 目录中创建文件
	targetCode := `use should_be_ignored::Something;`
	if err := os.WriteFile(filepath.Join(targetDir, "ignored.rs"), []byte(targetCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 在根目录创建文件
	rootCode := `use actual_dep::Something;`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.rs"), []byte(rootCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewRustAnalyzer()
	deps := []string{"actual_dep", "should_be_ignored"}

	result, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !result["actual_dep"].Used {
		t.Error("actual_dep should be marked as used")
	}
	if result["should_be_ignored"].Used {
		t.Error("should_be_ignored should not be marked as used (target dir should be skipped)")
	}
}

func TestExtractRustImports(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "Simple use",
			code:     `use serde::Deserialize;`,
			expected: []string{"serde::Deserialize"},
		},
		{
			name:     "Multiple use",
			code:     "use serde::Deserialize;\nuse tokio::sync::Mutex;",
			expected: []string{"serde::Deserialize", "tokio::sync::Mutex"},
		},
		{
			name:     "Extern crate",
			code:     `extern crate reqwest;`,
			expected: []string{"reqwest"},
		},
		{
			name:     "Mixed imports",
			code:     "use serde::Deserialize;\nextern crate tokio;\nuse reqwest::Client;",
			expected: []string{"serde::Deserialize", "tokio", "reqwest::Client"},
		},
		{
			name:     "With comments",
			code:     "// This is a comment\nuse serde::Deserialize;",
			expected: []string{"serde::Deserialize"},
		},
		{
			name:     "Block comment",
			code:     "/* comment */\nuse serde::Deserialize;",
			expected: []string{"serde::Deserialize"},
		},
		{
			name:     "String literal",
			code:     "let s = \"use fake::Something;\";\nuse real::Something;",
			expected: []string{"real::Something"},
		},
		{
			name:     "Empty code",
			code:     ``,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractRustImports(tt.code)

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

func TestResolveRustPackageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		importPath string
		expected   string
	}{
		{"Simple", "serde::Deserialize", "serde"},
		{"Deep path", "tokio::sync::Mutex", "tokio"},
		{"No path", "reqwest", "reqwest"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveRustPackageName(tt.importPath)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRemoveRustComments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comments",
			input:    `use serde::Deserialize;`,
			expected: `use serde::Deserialize;`,
		},
		{
			name:     "Line comment",
			input:    "// comment\nuse serde::Deserialize;",
			expected: "\nuse serde::Deserialize;",
		},
		{
			name:     "Block comment",
			input:    "/* comment */\nuse serde::Deserialize;",
			expected: "\nuse serde::Deserialize;",
		},
		{
			name:     "Comment in string",
			input:    `let s = "use fake::Something;";`,
			expected: `let s = "use fake::Something;";`,
		},
		{
			name:     "Empty",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeRustComments(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
