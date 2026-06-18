package surface

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SurfaceResult 影响面分析结果
type SurfaceResult struct {
	Package     string              // 包名
	Files       []string            // 使用该包的文件列表
	Modules     []string            // 使用的子模块/子路径列表
	RefCount    int                 // 总引用次数
	Criticality string              // 关键度：High/Medium/Low
	fileSet     map[string]struct{} // 用于文件去重
	moduleSet   map[string]struct{} // 用于模块去重
}

// 正则表达式匹配 import 语句
var (
	reImport        = regexp.MustCompile(`import\s+(?:[\w*{}\s,]+?\s+from\s+)?['"]([^'"]+)['"]`)
	reRequire       = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	reDynamicImport = regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

// AnalyzeSurface 分析依赖的影响面
func AnalyzeSurface(dir string, deps []string) (map[string]*SurfaceResult, error) {
	results := make(map[string]*SurfaceResult)

	// 初始化结果
	for _, dep := range deps {
		results[dep] = &SurfaceResult{
			Package:   dep,
			Files:     []string{},
			Modules:   []string{},
			fileSet:   make(map[string]struct{}),
			moduleSet: make(map[string]struct{}),
		}
	}

	// 构建包名集合用于快速查找
	depSet := make(map[string]bool)
	for _, dep := range deps {
		depSet[dep] = true
	}

	// 扫描所有源文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过 node_modules 和隐藏目录
		if info.IsDir() {
			name := info.Name()
			if name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// 只处理 JS/TS 文件
		ext := filepath.Ext(info.Name())
		if ext != ".js" && ext != ".ts" && ext != ".jsx" && ext != ".tsx" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)
		relPath, _ := filepath.Rel(dir, path)

		// 提取所有 import
		imports := extractImports(content)
		for _, imp := range imports {
			pkgName := resolvePackageName(imp)
			if depSet[pkgName] {
				result := results[pkgName]

				// 添加文件（使用 map 去重）
				if _, exists := result.fileSet[relPath]; !exists {
					result.fileSet[relPath] = struct{}{}
					result.Files = append(result.Files, relPath)
				}

				// 添加模块/子路径（使用 map 去重）
				if _, exists := result.moduleSet[imp]; !exists {
					result.moduleSet[imp] = struct{}{}
					result.Modules = append(result.Modules, imp)
				}

				// 增加引用计数
				result.RefCount++
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 计算关键度并排序
	for _, result := range results {
		result.Criticality = calculateCriticality(result)
		sort.Strings(result.Files)
		sort.Strings(result.Modules)
	}

	return results, nil
}

// extractImports 从源码中提取所有 import 路径
func extractImports(content string) []string {
	var imports []string

	// 匹配 import 语句
	matches := reImport.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// 匹配 require 语句
	matches = reRequire.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	// 匹配动态 import 语句
	matches = reDynamicImport.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// resolvePackageName 从 import 路径中提取包名
func resolvePackageName(importPath string) string {
	// 跳过相对路径
	if strings.HasPrefix(importPath, ".") {
		return ""
	}

	// 处理 scoped packages (@org/pkg)
	if strings.HasPrefix(importPath, "@") {
		parts := strings.Split(importPath, "/")
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return importPath
	}

	// 普通包：取第一段
	parts := strings.Split(importPath, "/")
	return parts[0]
}

// calculateCriticality 计算依赖的关键度
func calculateCriticality(result *SurfaceResult) string {
	fileCount := len(result.Files)
	refCount := result.RefCount

	// 关键度评分标准：
	// High: 10+ 文件 或 50+ 引用
	// Medium: 3-9 文件 或 10-49 引用
	// Low: 1-2 文件 或 1-9 引用

	if fileCount >= 10 || refCount >= 50 {
		return "High"
	} else if fileCount >= 3 || refCount >= 10 {
		return "Medium"
	}
	return "Low"
}
