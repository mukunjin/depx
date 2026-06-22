package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mukunjin/depx/internal/config"
	"github.com/mukunjin/depx/internal/lockfile"
	"github.com/mukunjin/depx/internal/manifest"
	"github.com/mukunjin/depx/internal/usage"
)

// typesPrefix 是 TypeScript 类型包的统一前缀
const typesPrefix = "@types/"

// ScanResult 扫描结果
type ScanResult struct {
	Path         string                           // 项目路径
	ManifestType string                           // 包管理器类型
	TotalDeps    int                              // 总依赖数
	UsedDeps     int                              // 已使用依赖数
	UnusedDeps   int                              // 未使用依赖数
	TypePackages int                              // 类型包数量（@types/xxx）
	UsageDetails map[string]*manifest.UsageResult // 每个依赖的使用详情
	IndirectDeps []string                         // 间接依赖列表（来自 lockfile）
	TypePkgNames []string                         // 类型包名称列表
	
	// 新增：区分 runtime 和 tool packages
	RuntimeDeps []string // 运行时依赖
	ToolDeps    []string // 工具包（devDependencies 中的包）
	
	// 新增：间接依赖共享计数
	IndirectSharedCounts map[string]int // 间接依赖被多少个直接依赖引用
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

	// 获取依赖列表（scan 同时检查 runtime 与 dev 依赖）
	runtimeDeps, err := m.Dependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to read dependencies: %w", err)
	}
	
	devDeps, err := m.DevDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to read dev dependencies: %w", err)
	}
	
	allDeps := append(runtimeDeps, devDeps...)

	// 过滤被忽略的依赖
	filteredDeps := make([]string, 0, len(allDeps))
	for _, dep := range allDeps {
		if !cfg.IsIgnored(dep) {
			filteredDeps = append(filteredDeps, dep)
		}
	}

	// 构建分析器选项
	opts := &usage.Options{
		ExcludeDirs:     cfg.ExcludeDirs,
		ExcludeFiles:    cfg.ExcludeFiles,
		ReadNodeModules: cfg.ReadNodeModules,
	}

	// 选择合适的分析器
	analyzer, err := selectAnalyzer(m.Type())
	if err != nil {
		return nil, err
	}

	// 分析使用情况
	usageMap, err := analyzer.Analyze(absDir, filteredDeps, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze usage: %w", err)
	}

	// 统计结果
	result := &ScanResult{
		Path:         absDir,
		ManifestType: m.Type(),
		UsageDetails: usageMap,
		RuntimeDeps:  make([]string, 0),
		ToolDeps:     make([]string, 0),
	}

	// 创建 devDependencies 集合用于快速查找
	devDepsMap := make(map[string]bool)
	for _, dep := range devDeps {
		devDepsMap[dep] = true
	}

	for pkg, r := range usageMap {
		// 类型包（@types/xxx）完全独立统计，不计入 TotalDeps/UsedDeps/UnusedDeps
		if strings.HasPrefix(pkg, typesPrefix) {
			result.TypePackages++
			result.TypePkgNames = append(result.TypePkgNames, pkg)
			continue
		}
		
		// 区分 runtime 和 tool packages
		if devDepsMap[pkg] {
			result.ToolDeps = append(result.ToolDeps, pkg)
		} else {
			result.RuntimeDeps = append(result.RuntimeDeps, pkg)
		}
		
		result.TotalDeps++
		if r.Used {
			result.UsedDeps++
		} else {
			result.UnusedDeps++
		}
	}

	// 如果启用 lockfile 分析，获取间接依赖
	if cfg.LockFile {
		lf, err := lockfile.DetectLockFile(absDir)
		if err != nil {
			// lockfile 不存在或检测失败是正常情况，记录警告但不中断
			fmt.Fprintf(os.Stderr, "Warning: could not detect lockfile: %v\n", err)
		} else {
			lockDeps, err := lf.Dependencies()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not read lockfile dependencies: %v\n", err)
			} else if len(lockDeps) == 0 {
				// 空依赖列表，跳过处理
			} else {
				// 提取间接依赖（在 lockfile 中但不在 manifest 中的依赖）
				manifestDeps := make(map[string]bool)
				for _, dep := range allDeps {
					manifestDeps[dep] = true
				}

				indirectSet := make(map[string]bool)
				for _, ld := range lockDeps {
					if strings.HasPrefix(ld.Name, typesPrefix) {
						continue
					}
					if !manifestDeps[ld.Name] && !indirectSet[ld.Name] {
						indirectSet[ld.Name] = true
						result.IndirectDeps = append(result.IndirectDeps, ld.Name)
					}
				}
				
				// 计算间接依赖的共享计数（被多少个直接依赖引用）
				if npmLf, ok := lf.(*lockfile.NpmLockFile); ok {
					result.IndirectSharedCounts = npmLf.SharedIndirectCounts(allDeps)
				}
			}
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
