package filter

import (
	"testing"
)

func TestShouldExcludeFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{
			name:     "空路径",
			path:     "",
			patterns: []string{"*.log"},
			want:     false,
		},
		{
			name:     "空模式列表",
			path:     "test.log",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "空模式项",
			path:     "test.log",
			patterns: []string{""},
			want:     false,
		},
		{
			name:     "通配符匹配 - *.log",
			path:     "test.log",
			patterns: []string{"*.log"},
			want:     true,
		},
		{
			name:     "通配符匹配 - *.txt",
			path:     "/path/to/file.txt",
			patterns: []string{"*.txt"},
			want:     true,
		},
		{
			name:     "通配符不匹配",
			path:     "test.go",
			patterns: []string{"*.log"},
			want:     false,
		},
		{
			name:     "路径包含匹配",
			path:     "/path/to/node_modules/file.js",
			patterns: []string{"node_modules"},
			want:     true,
		},
		{
			name:     "路径包含匹配 - 多个模式",
			path:     "/path/to/vendor/lib.go",
			patterns: []string{"node_modules", "vendor"},
			want:     true,
		},
		{
			name:     "Windows 路径兼容性",
			path:     `C:\path\to\node_modules\file.js`,
			patterns: []string{"node_modules"},
			want:     true,
		},
		{
			name:     "多个模式 - 第一个匹配",
			path:     "test.log",
			patterns: []string{"*.log", "*.tmp"},
			want:     true,
		},
		{
			name:     "多个模式 - 第二个匹配",
			path:     "test.tmp",
			patterns: []string{"*.log", "*.tmp"},
			want:     true,
		},
		{
			name:     "多个模式 - 都不匹配",
			path:     "test.go",
			patterns: []string{"*.log", "*.tmp"},
			want:     false,
		},
		{
			name:     "混合空模式和有效模式",
			path:     "test.log",
			patterns: []string{"", "*.log", ""},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExcludeFile(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("ShouldExcludeFile(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestShouldExcludeDir(t *testing.T) {
	tests := []struct {
		name        string
		dir         string
		excludeDirs []string
		want        bool
	}{
		{
			name:        "空目录名",
			dir:         "",
			excludeDirs: []string{"node_modules"},
			want:        false,
		},
		{
			name:        "空排除列表",
			dir:         "node_modules",
			excludeDirs: []string{},
			want:        false,
		},
		{
			name:        "空排除项",
			dir:         "node_modules",
			excludeDirs: []string{""},
			want:        false,
		},
		{
			name:        "精确匹配 - node_modules",
			dir:         "node_modules",
			excludeDirs: []string{"node_modules"},
			want:        true,
		},
		{
			name:        "精确匹配 - vendor",
			dir:         "vendor",
			excludeDirs: []string{"node_modules", "vendor"},
			want:        true,
		},
		{
			name:        "不匹配",
			dir:         "src",
			excludeDirs: []string{"node_modules", "vendor"},
			want:        false,
		},
		{
			name:        "多个排除目录 - 第一个匹配",
			dir:         "node_modules",
			excludeDirs: []string{"node_modules", "vendor", "dist"},
			want:        true,
		},
		{
			name:        "多个排除目录 - 中间匹配",
			dir:         "vendor",
			excludeDirs: []string{"node_modules", "vendor", "dist"},
			want:        true,
		},
		{
			name:        "多个排除目录 - 最后匹配",
			dir:         "dist",
			excludeDirs: []string{"node_modules", "vendor", "dist"},
			want:        true,
		},
		{
			name:        "混合空排除项和有效排除项",
			dir:         "vendor",
			excludeDirs: []string{"", "vendor", ""},
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExcludeDir(tt.dir, tt.excludeDirs)
			if got != tt.want {
				t.Errorf("ShouldExcludeDir(%q, %v) = %v, want %v", tt.dir, tt.excludeDirs, got, tt.want)
			}
		})
	}
}
