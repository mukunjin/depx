package usage

import (
	"github.com/mukunjin/depx/internal/filter"
	"github.com/mukunjin/depx/internal/manifest"
)

// Options 定义分析器选项
type Options struct {
	// ExcludeDirs 排除的目录
	ExcludeDirs []string
	// ExcludeFiles 排除的文件模式
	ExcludeFiles []string
	// ReadNodeModules 是否读取 node_modules
	ReadNodeModules bool
}

// Analyzer 定义依赖使用分析器接口
type Analyzer interface {
	// Analyze 扫描目录下所有源文件，返回每个依赖的使用情况
	Analyze(dir string, deps []string, opts *Options) (map[string]*manifest.UsageResult, error)
}

// shouldExcludeFile 检查文件是否应该被排除
func shouldExcludeFile(path string, patterns []string) bool {
	return filter.ShouldExcludeFile(path, patterns)
}
