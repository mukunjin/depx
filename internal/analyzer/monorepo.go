package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MonorepoResult Monorepo 扫描结果
type MonorepoResult struct {
	Path             string        // 项目根路径
	Workspaces       []string      // 工作区列表
	WorkspaceResults []*ScanResult // 每个工作区的扫描结果
	TotalDeps        int           // 总依赖数（去重后）
	UsedDeps         int           // 已使用依赖数
	UnusedDeps       int           // 未使用依赖数
}

// detectWorkspaces 检测项目是否为 monorepo，返回工作区列表
func detectWorkspaces(dir string) ([]string, error) {
	// 尝试检测 npm workspaces
	pkgPath := filepath.Join(dir, "package.json")
	if _, err := os.Stat(pkgPath); err == nil {
		data, err := os.ReadFile(pkgPath)
		if err != nil {
			return nil, err
		}

		var pkg struct {
			Workspaces []string `json:"workspaces"`
		}
		if err := json.Unmarshal(data, &pkg); err != nil {
			return nil, err
		}

		if len(pkg.Workspaces) > 0 {
			return expandWorkspacePatterns(dir, pkg.Workspaces)
		}
	}

	// 未来可以添加其他 monorepo 格式的支持
	// - yarn workspaces (与 npm 相同)
	// - pnpm workspaces (pnpm-workspace.yaml)
	// - go workspaces (go.work)

	return nil, nil
}

// expandWorkspacePatterns 展开工作区模式（如 "packages/*"）
func expandWorkspacePatterns(rootDir string, patterns []string) ([]string, error) {
	var workspaces []string

	for _, pattern := range patterns {
		if isGlobPattern(pattern) {
			// 展开 glob 模式
			fullPattern := filepath.Join(rootDir, pattern)
			matches, err := filepath.Glob(fullPattern)
			if err != nil {
				return nil, fmt.Errorf("failed to expand pattern %s: %w", pattern, err)
			}

			for _, match := range matches {
				if info, err := os.Stat(match); err == nil && info.IsDir() {
					pkgPath := filepath.Join(match, "package.json")
					if _, err := os.Stat(pkgPath); err == nil {
						workspaces = append(workspaces, match)
					}
				}
			}
		} else {
			// 直接路径
			fullPath := filepath.Join(rootDir, pattern)
			if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
				workspaces = append(workspaces, fullPath)
			}
		}
	}

	return workspaces, nil
}

// isGlobPattern 检查字符串是否包含通配符
func isGlobPattern(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

// ScanMonorepo 扫描 monorepo 项目
func ScanMonorepo(dir string) (*MonorepoResult, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	// 检测工作区
	workspaces, err := detectWorkspaces(absDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		return nil, fmt.Errorf("no workspaces found")
	}

	result := &MonorepoResult{
		Path:       absDir,
		Workspaces: workspaces,
	}

	// 扫描每个工作区
	allDeps := make(map[string]bool)
	usedDeps := make(map[string]bool)

	for _, ws := range workspaces {
		wsResult, err := Scan(ws)
		if err != nil {
			// 跳过无法扫描的工作区
			continue
		}

		result.WorkspaceResults = append(result.WorkspaceResults, wsResult)

		// 合并依赖统计
		for pkg, usage := range wsResult.UsageDetails {
			allDeps[pkg] = true
			if usage.Used {
				usedDeps[pkg] = true
			}
		}
	}

	result.TotalDeps = len(allDeps)
	result.UsedDeps = len(usedDeps)
	result.UnusedDeps = result.TotalDeps - result.UsedDeps

	return result, nil
}
