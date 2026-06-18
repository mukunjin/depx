package usage

import "github.com/depx/depx/internal/manifest"

// Analyzer 定义依赖使用分析器接口
type Analyzer interface {
	// Analyze 扫描目录下所有源文件，返回每个依赖的使用情况
	Analyze(dir string, deps []string) (map[string]*manifest.UsageResult, error)
}
