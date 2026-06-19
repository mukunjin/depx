package usage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoAnalyzer(t *testing.T) {
	t.Parallel()
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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Expected dependency '%s' to be used", dep)
		}
	}
}

func TestGoAnalyzerSubpath(t *testing.T) {
	t.Parallel()
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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["github.com/go-redis/redis/v8"].Used {
		t.Error("Expected 'github.com/go-redis/redis/v8' to be used")
	}
}

func TestGoAnalyzerSkipDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results["github.com/some/pkg"].Used {
		t.Error("Expected 'github.com/some/pkg' to not be used (should skip vendor)")
	}
}

func TestGoAnalyzerMultiFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	mainGo := `package main

import "github.com/gin-gonic/gin"

func main() {
	_ = gin.Default()
}
`
	handlerGo := `package handlers

import "github.com/spf13/cobra"

func Run() {
	_ = cobra.Command{}
}
`
	subDir := filepath.Join(tmpDir, "handlers")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "handler.go"), []byte(handlerGo), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	deps := []string{"github.com/gin-gonic/gin", "github.com/spf13/cobra", "github.com/unused/pkg"}

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["github.com/gin-gonic/gin"].Used {
		t.Error("gin should be used")
	}
	if !results["github.com/spf13/cobra"].Used {
		t.Error("cobra should be used")
	}
	if results["github.com/unused/pkg"].Used {
		t.Error("unused pkg should not be used")
	}

	// 验证 UsedIn 字段
	if len(results["github.com/gin-gonic/gin"].UsedIn) != 1 {
		t.Errorf("gin should be used in 1 file, got %d", len(results["github.com/gin-gonic/gin"].UsedIn))
	}
	if len(results["github.com/spf13/cobra"].UsedIn) != 1 {
		t.Errorf("cobra should be used in 1 file, got %d", len(results["github.com/spf13/cobra"].UsedIn))
	}
}

func TestGoAnalyzerTestFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	mainGo := `package main

import "github.com/gin-gonic/gin"

func main() {
	_ = gin.Default()
}
`
	testGo := `package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {
	assert.True(t, true)
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testGo), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	deps := []string{"github.com/gin-gonic/gin", "github.com/stretchr/testify"}

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["github.com/gin-gonic/gin"].Used {
		t.Error("gin should be used")
	}
	if !results["github.com/stretchr/testify"].Used {
		t.Error("testify should be used in test file")
	}
}

func TestGoAnalyzerNonExistentDir(t *testing.T) {
	t.Parallel()
	analyzer := NewGoAnalyzer()
	_, err := analyzer.Analyze("/non/existent/dir", []string{"fmt"}, nil)
	if err == nil {
		t.Error("Should return error for non-existent directory")
	}
}

func TestGoAnalyzerEmptyDeps(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	goCode := `package main
import "fmt"
func main() { fmt.Println("hi") }
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(goCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewGoAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{}, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestExtractGoImports(t *testing.T) {
	t.Parallel()
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
		{
			name: "Dot import",
			code: `package main

import . "fmt"
`,
			expected: []string{"fmt"},
		},
		{
			name: "Blank import",
			code: `package main

import _ "image/png"
`,
			expected: []string{"image/png"},
		},
		{
			name: "Import with inline comment",
			code: `package main

import (
	"fmt" // standard library
	"github.com/gin-gonic/gin" // web framework
)
`,
			expected: []string{"fmt", "github.com/gin-gonic/gin"},
		},
		{
			name: "Import block without space",
			code: `package main

import(
	"fmt"
	"os"
)
`,
			expected: []string{"fmt", "os"},
		},
		{
			name: "Import block with empty lines",
			code: `package main

import (
	"fmt"

	"os"

	"strings"
)
`,
			expected: []string{"fmt", "os", "strings"},
		},
		{
			name: "Full line comment inside import block",
			code: `package main

import (
	"fmt"
	// "os"
	"strings"
)
`,
			expected: []string{"fmt", "strings"},
		},
		{
			name: "URL in import path",
			code: `package main

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)
`,
			expected: []string{"github.com/gin-gonic/gin", "golang.org/x/text/language"},
		},
		{
			name: "Versioned import path",
			code: `package main

import (
	"github.com/go-redis/redis/v8"
)
`,
			expected: []string{"github.com/go-redis/redis/v8"},
		},
		{
			name: "Multiple single imports",
			code: `package main

import "fmt"
import "os"
import "strings"
`,
			expected: []string{"fmt", "os", "strings"},
		},
		{
			name: "Mixed single and block imports",
			code: `package main

import "fmt"

import (
	"os"
	"strings"
)
`,
			expected: []string{"fmt", "os", "strings"},
		},
		{
			name: "No imports",
			code: `package main

func main() {}
`,
			expected: nil,
		},
		{
			name:     "Empty string",
			code:     "",
			expected: nil,
		},
		{
			name: "Import with grouped stdlib and third-party",
			code: `package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)
`,
			expected: []string{"fmt", "os", "github.com/gin-gonic/gin", "github.com/spf13/cobra"},
		},
		{
			name: "Import with aliased and blank",
			code: `package main

import (
	"fmt"
	myfmt "fmt"
	_ "image/png"
)
`,
			expected: []string{"fmt", "fmt", "image/png"},
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
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple", `"fmt"`, "fmt"},
		{"Aliased", `f "fmt"`, "fmt"},
		{"Dot import", `. "fmt"`, "fmt"},
		{"Blank import", `_ "image/png"`, "image/png"},
		{"With inline comment", `"fmt" // comment`, "fmt"},
		{"URL path", `"github.com/gin-gonic/gin"`, "github.com/gin-gonic/gin"},
		{"URL path with comment", `"github.com/gin-gonic/gin" // web framework`, "github.com/gin-gonic/gin"},
		{"Versioned path", `"github.com/go-redis/redis/v8"`, "github.com/go-redis/redis/v8"},
		{"Empty", "", ""},
		{"Only whitespace", "   ", ""},
		{"No quotes", "nopackage", ""},
		{"Unclosed quote", `"unclosed`, ""},
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

func TestRemoveGoLineComment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comment",
			input:    `"fmt"`,
			expected: `"fmt"`,
		},
		{
			name:     "Simple inline comment",
			input:    `"fmt" // standard library`,
			expected: `"fmt"`,
		},
		{
			name:     "URL path with comment",
			input:    `"github.com/gin-gonic/gin" // web framework`,
			expected: `"github.com/gin-gonic/gin"`,
		},
		{
			name:     "No string no comment",
			input:    `import (`,
			expected: `import (`,
		},
		{
			name:     "No string with comment",
			input:    `// this is a comment`,
			expected: ``,
		},
		{
			name:     "Aliased with comment",
			input:    `f "fmt" // alias`,
			expected: `f "fmt"`,
		},
		{
			name:     "Versioned path with comment",
			input:    `"github.com/go-redis/redis/v8" // redis client`,
			expected: `"github.com/go-redis/redis/v8"`,
		},
		{
			name:     "Empty string",
			input:    ``,
			expected: ``,
		},
		{
			name:     "Only string",
			input:    `"fmt"`,
			expected: `"fmt"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeGoLineComment(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
