package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/depx/depx/internal/manifest"
	"github.com/depx/depx/internal/usage"
)

// ScanResult 扫描结果
type ScanResult struct {
	Path           string                          // 项目路径
	ManifestType   string                          // 包管理器类型
	TotalDeps      int                             // 总依赖数
	UsedDeps       int                             // 已使用依赖数
	UnusedDeps     int                             // 未使用依赖数
	UsageDetails   map[string]*manifest.UsageResult // 每个依赖的使用详情
}

// Scan 扫描项目，检测未使用的依赖
func Scan(dir string) (*ScanResult, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// 尝试检测项目类型
	m, err := detectManifest(absDir)
	if err != nil {
		return nil, err
	}

	// 获取依赖列表
	deps, err := m.Dependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to read dependencies: %w", err)
	}

	// 选择合适的分析器
	analyzer, err := selectAnalyzer(m.Type())
	if err != nil {
		return nil, err
	}

	// 分析使用情况
	usageMap, err := analyzer.Analyze(absDir, deps)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze usage: %w", err)
	}

	// 统计结果
	result := &ScanResult{
		Path:         absDir,
		ManifestType: m.Type(),
		TotalDeps:    len(deps),
		UsageDetails: usageMap,
	}

	for _, r := range usageMap {
		if r.Used {
			result.UsedDeps++
		} else {
			result.UnusedDeps++
		}
	}

	return result, nil
}

// detectManifest 检测项目类型
func detectManifest(dir string) (manifest.Manifest, error) {
	// 优先检测 npm
	if npm, err := manifest.NewNpmManifest(dir); err == nil {
		return npm, nil
	}

	// 检测 go
	if gomod, err := manifest.NewGoModManifest(dir); err == nil {
		return gomod, nil
	}

	return nil, fmt.Errorf("no supported project found (package.json or go.mod)")
}

// selectAnalyzer 根据项目类型选择分析器
func selectAnalyzer(manifestType string) (usage.Analyzer, error) {
	switch manifestType {
	case "npm":
		return usage.NewJSAnalyzer(), nil
	case "go":
		return usage.NewGoAnalyzer(), nil
	default:
		return nil, fmt.Errorf("unsupported manifest type: %s", manifestType)
	}
}
