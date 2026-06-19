package filter

import (
	"path/filepath"
	"testing"
)

func TestShouldExcludeFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{name: "empty path", path: "", patterns: []string{"*.go"}, want: false},
		{name: "no patterns", path: "src/main.go", patterns: nil, want: false},
		{name: "base match", path: "src/foo_test.go", patterns: []string{"foo_test.go"}, want: true},
		{name: "glob match", path: "pkg/file.go", patterns: []string{"*.go"}, want: true},
		{name: "contains pattern (windows style)", path: filepath.FromSlash("C:/project/dist/file.js"), patterns: []string{"dist"}, want: true},
		{name: "contains pattern with slash", path: filepath.FromSlash("/home/user/project/src/file.js"), patterns: []string{"src/"}, want: true},
		{name: "non matching pattern", path: "main.py", patterns: []string{"*.js"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExcludeFile(tt.path, tt.patterns)
			if got != tt.want {
				t.Fatalf("ShouldExcludeFile(%q, %v) = %v; want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestShouldExcludeDir(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		dir         string
		excludeDirs []string
		want        bool
	}{
		{name: "empty dir", dir: "", excludeDirs: []string{"vendor"}, want: false},
		{name: "direct match", dir: "vendor", excludeDirs: []string{"vendor", "dist"}, want: true},
		{name: "no match", dir: "src", excludeDirs: []string{"vendor"}, want: false},
		{name: "path not equal to simple name", dir: filepath.FromSlash("C:/project/vendor"), excludeDirs: []string{"vendor"}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExcludeDir(tt.dir, tt.excludeDirs)
			if got != tt.want {
				t.Fatalf("ShouldExcludeDir(%q, %v) = %v; want %v", tt.dir, tt.excludeDirs, got, tt.want)
			}
		})
	}
}
