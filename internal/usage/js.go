package usage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/depx/depx/internal/manifest"
)

// JSAnalyzer 分析 JS/TS 文件中的 import 使用
type JSAnalyzer struct{}

func NewJSAnalyzer() *JSAnalyzer { return &JSAnalyzer{} }

// 匹配的源文件扩展名
var jsExtensions = map[string]bool{
	".js":     true,
	".ts":     true,
	".jsx":    true,
	".tsx":    true,
	".mjs":    true,
	".cjs":    true,
	".vue":    true,
	".svelte": true,
}

// 跳过的目录
var jsSkipDirs = map[string]bool{
	"node_modules": true,
	"dist":         true,
	"build":        true,
	".git":         true,
	".next":        true,
	"coverage":     true,
	".nuxt":        true,
}

// import 匹配正则
var (
	// import ... from 'package' 或 import 'package'
	reESMImport = regexp.MustCompile(`(?m)import\s+(?:[\w*{}\s,]+?\s+from\s+)?['"]([^'"]+)['"]`)
	// require('package')
	reRequire = regexp.MustCompile(`(?m)require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	// 动态 import('package')
	reDynamicImport = regexp.MustCompile(`(?m)import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	// export ... from 'package'
	reExportFrom = regexp.MustCompile(`(?m)export\s+(?:[\w*{}\s,]+?\s+from\s+)['"]([^'"]+)['"]`)
)

func (j *JSAnalyzer) Analyze(dir string, deps []string) (map[string]*manifest.UsageResult, error) {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	result := make(map[string]*manifest.UsageResult)
	for _, dep := range deps {
		result[dep] = &manifest.UsageResult{Package: dep}
	}

	// 构建包名集合用于快速查找
	depSet := make(map[string]bool, len(deps))
	for _, dep := range deps {
		depSet[dep] = true
	}

	// 为每个依赖创建文件追踪 map
	fileTrackers := make(map[string]map[string]bool, len(deps))
	for _, dep := range deps {
		fileTrackers[dep] = make(map[string]bool)
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		if info.IsDir() {
			if jsSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(info.Name())
		if !jsExtensions[ext] {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		relPath, _ := filepath.Rel(dir, path)

		// 提取所有 import 的包名
		imports := extractJSImports(content)
		for _, imp := range imports {
			pkgName := resolveJSPackageName(imp)
			if depSet[pkgName] {
				r := result[pkgName]
				r.Used = true
				r.RefCount++

				// 使用 map 追踪文件，避免重复添加
				tracker := fileTrackers[pkgName]
				if !tracker[relPath] {
					tracker[relPath] = true
					r.UsedIn = append(r.UsedIn, relPath)
				}
			}
		}

		return nil
	})

	return result, err
}

// removeComments 移除代码中的注释，正确处理字符串字面量
func removeComments(content string) string {
	var result strings.Builder
	result.Grow(len(content))

	i := 0
	n := len(content)

	for i < n {
		// 检查字符串字面量
		if content[i] == '"' || content[i] == '\'' || content[i] == '`' {
			quote := content[i]
			result.WriteByte(content[i])
			i++

			// 读取整个字符串
			for i < n {
				if content[i] == '\\' && i+1 < n {
					// 转义字符
					result.WriteByte(content[i])
					result.WriteByte(content[i+1])
					i += 2
					continue
				}
				if content[i] == quote {
					result.WriteByte(content[i])
					i++
					break
				}
				result.WriteByte(content[i])
				i++
			}
			continue
		}

		// 检查单行注释
		if i+1 < n && content[i] == '/' && content[i+1] == '/' {
			// 跳过直到行尾
			for i < n && content[i] != '\n' {
				i++
			}
			continue
		}

		// 检查多行注释
		if i+1 < n && content[i] == '/' && content[i+1] == '*' {
			i += 2
			// 跳过直到 */
			for i+1 < n {
				if content[i] == '*' && content[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			continue
		}

		// 普通字符
		result.WriteByte(content[i])
		i++
	}

	return result.String()
}

// extractJSImports 从源码中提取所有 import 的模块路径
func extractJSImports(content string) []string {
	var imports []string

	// 移除注释以避免误识别
	cleanContent := removeComments(content)

	for _, match := range reESMImport.FindAllStringSubmatch(cleanContent, -1) {
		imports = append(imports, match[1])
	}
	for _, match := range reRequire.FindAllStringSubmatch(cleanContent, -1) {
		imports = append(imports, match[1])
	}
	for _, match := range reDynamicImport.FindAllStringSubmatch(cleanContent, -1) {
		imports = append(imports, match[1])
	}
	for _, match := range reExportFrom.FindAllStringSubmatch(cleanContent, -1) {
		imports = append(imports, match[1])
	}

	return imports
}

// resolveJSPackageName 从 import path 中提取包名
// 例如: "lodash/get" -> "lodash", "@org/pkg/sub" -> "@org/pkg"
func resolveJSPackageName(importPath string) string {
	// 跳过相对路径
	if strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/") {
		return ""
	}

	parts := strings.Split(importPath, "/")
	if len(parts) == 0 {
		return ""
	}

	// scoped package: @org/pkg
	if strings.HasPrefix(parts[0], "@") && len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}

	return parts[0]
}
