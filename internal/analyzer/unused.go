package analyzer

import (
	"fmt"
	"path/filepath"

	"github.com/mukunjin/depx/internal/config"
	"github.com/mukunjin/depx/internal/manifest"
	"github.com/mukunjin/depx/internal/usage"
)

// ScanResult 扫描结果
type ScanResult struct {
	Path         string                           // 项目路径
	ManifestType string                           // 包管理器类型
	TotalDeps    int                              // 总依赖数
	UsedDeps     int                              // 已使用依赖数
	UnusedDeps   int                              // 未使用依赖数
	UsageDetails map[string]*manifest.UsageResult // 每个依赖的使用详情
}

// Scan 扫描项目，检测未使用的依赖
func Scan(dir string) (*ScanResult, error) {
	return ScanWithConfig(dir, nil)
}

// ScanWithConfig 使用配置文件扫描项目
func ScanWithConfig(dir string, cfg *config.Config) (*ScanResult, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// 加载配置
	if cfg == nil {
		cfg, err = config.FindAndLoad(absDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	// 尝试检测项目类型
	m, err := DetectManifest(absDir)
	if err != nil {
		return nil, err
	}

	// 获取依赖列表
	deps, err := m.Dependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to read dependencies: %w", err)
	}

	// 过滤被忽略的依赖
	filteredDeps := make([]string, 0, len(deps))
	for _, dep := range deps {
		if !cfg.IsIgnored(dep) {
			filteredDeps = append(filteredDeps, dep)
		}
	}

	// 选择合适的分析器
	analyzer, err := selectAnalyzer(m.Type())
	if err != nil {
		return nil, err
	}

	// 分析使用情况
	usageMap, err := analyzer.Analyze(absDir, filteredDeps)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze usage: %w", err)
	}

	// 统计结果
	result := &ScanResult{
		Path:         absDir,
		ManifestType: m.Type(),
		TotalDeps:    len(filteredDeps),
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

// DetectManifest 检测项目类型
func DetectManifest(dir string) (manifest.Manifest, error) {
	// 优先检测 npm
	if npm, err := manifest.NewNpmManifest(dir); err == nil {
		return npm, nil
	}

	// 检测 go
	if gomod, err := manifest.NewGoModManifest(dir); err == nil {
		return gomod, nil
	}

	// 检测 cargo
	if cargo, err := manifest.NewCargoManifest(dir); err == nil {
		return cargo, nil
	}

	// 检测 pip
	if pip, err := manifest.NewPipManifest(dir); err == nil {
		return pip, nil
	}

	return nil, fmt.Errorf("no supported project found (package.json, go.mod, Cargo.toml, or requirements.txt)")
}

// selectAnalyzer 根据项目类型选择分析器
func selectAnalyzer(manifestType string) (usage.Analyzer, error) {
	switch manifestType {
	case "npm":
		return usage.NewJSAnalyzer(), nil
	case "go":
		return usage.NewGoAnalyzer(), nil
	case "cargo":
		return usage.NewRustAnalyzer(), nil
	case "pip":
		return usage.NewPythonAnalyzer(), nil
	default:
		return nil, fmt.Errorf("unsupported manifest type: %s", manifestType)
	}
}
