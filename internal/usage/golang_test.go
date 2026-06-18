package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()

	goCode := `package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	alias "github.com/spf13/cobra"
)

func main() {
	fmt.Println("test")
	_ = gin.Default()
	_ = alias.Command{}
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	deps := []string{"github.com/gin-gonic/gin", "github.com/spf13/cobra"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 验证所有依赖都被检测到
	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Expected dependency '%s' to be used", dep)
		}
	}
}

func TestGoAnalyzerSubpath(t *testing.T) {
	tmpDir := t.TempDir()

	goCode := `package main

import (
	"github.com/go-redis/redis/v8"
)

func main() {
	_ = redis.NewClient(nil)
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	deps := []string{"github.com/go-redis/redis/v8"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["github.com/go-redis/redis/v8"].Used {
		t.Error("Expected 'github.com/go-redis/redis/v8' to be used")
	}
}

func TestGoAnalyzerSkipDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// 在 vendor 目录中创建文件（应该被跳过）
	vendorDir := filepath.Join(tmpDir, "vendor", "github.com", "unused")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatal(err)
	}

	goCode := `package unused

import "github.com/some/pkg"
`
	if err := os.WriteFile(filepath.Join(vendorDir, "unused.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	deps := []string{"github.com/some/pkg"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// vendor 目录中的导入应该被忽略
	if results["github.com/some/pkg"].Used {
		t.Error("Expected 'github.com/some/pkg' to not be used (should skip vendor)")
	}
}

func TestExtractGoImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "Single import",
			code: `package main
import "fmt"
`,
			expected: []string{"fmt"},
		},
		{
			name: "Multiple imports",
			code: `package main

import (
	"fmt"
	"os"
)
`,
			expected: []string{"fmt", "os"},
		},
		{
			name: "Aliased import",
			code: `package main

import f "fmt"
`,
			expected: []string{"fmt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractGoImports(tt.code)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("Expected import[%d] = '%s', got '%s'", i, exp, result[i])
				}
			}
		})
	}
}

func TestParseGoImportLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple", `"fmt"`, "fmt"},
		{"Aliased", `f "fmt"`, "fmt"},
		{"With comment", `"fmt" // comment`, "fmt"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGoImportLine(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
