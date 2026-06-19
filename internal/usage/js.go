package usage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mukunjin/depx/internal/manifest"
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

func (j *JSAnalyzer) Analyze(dir string, deps []string, opts *Options) (map[string]*manifest.UsageResult, error) {
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

	// 构建跳过目录集合
	skipDirs := make(map[string]bool)
	for k, v := range jsSkipDirs {
		skipDirs[k] = v
	}
	// 如果启用 ReadNodeModules，则不跳过 node_modules
	if opts != nil && opts.ReadNodeModules {
		delete(skipDirs, "node_modules")
	}
	// 添加自定义排除目录
	if opts != nil {
		for _, d := range opts.ExcludeDirs {
			skipDirs[d] = true
		}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 返回错误而不是静默跳过
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !jsExtensions[ext] {
			return nil
		}

		// 检查文件模式排除
		if opts != nil && shouldExcludeFile(path, opts.ExcludeFiles) {
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

// removeComments 移除代码中的注释，保留字符串字面量内容
func removeComments(content string) string {
	var result strings.Builder
	result.Grow(len(content))

	i, n := 0, len(content)
	for i < n {
		c := content[i]

		// 处理字符串字面量：完整保留
		if c == '"' || c == '\'' || c == '`' {
			result.WriteByte(c)
			i++
			for i < n {
				if content[i] == '\\' && i+1 < n {
					result.WriteByte(content[i])
					result.WriteByte(content[i+1])
					i += 2
					continue
				}
				result.WriteByte(content[i])
				if content[i] == c {
					i++
					break
				}
				i++
			}
			continue
		}

		// 处理注释
		if c == '/' && i+1 < n {
			if content[i+1] == '/' {
				i = skipLineComment(content, i, n)
				continue
			}
			if content[i+1] == '*' {
				i = skipBlockComment(content, i, n)
				continue
			}
		}

		result.WriteByte(c)
		i++
	}

	return result.String()
}

// buildStringMask 构建字符串区域掩码，标记哪些位置在字符串字面量内部
func buildStringMask(content string) []bool {
	mask := make([]bool, len(content))
	i, n := 0, len(content)

	for i < n {
		c := content[i]

		if c == '"' || c == '\'' || c == '`' {
			quote := c
			mask[i] = true // 开始引号
			i++
			for i < n {
				mask[i] = true
				if content[i] == '\\' && i+1 < n {
					mask[i+1] = true
					i += 2
					continue
				}
				if content[i] == quote {
					i++
					break
				}
				i++
			}
			continue
		}

		// 跳过注释
		if c == '/' && i+1 < n {
			if content[i+1] == '/' {
				for i < n && content[i] != '\n' {
					mask[i] = true
					i++
				}
				continue
			}
			if content[i+1] == '*' {
				i += 2
				for i+1 < n {
					if content[i] == '*' && content[i+1] == '/' {
						i += 2
						break
					}
					i++
				}
				continue
			}
		}

		i++
	}

	return mask
}

// skipLineComment 跳过单行注释
func skipLineComment(content string, i, n int) int {
	for i < n && content[i] != '\n' {
		i++
	}
	return i
}

// skipBlockComment 跳过多行注释
func skipBlockComment(content string, i, n int) int {
	i += 2
	for i+1 < n {
		if content[i] == '*' && content[i+1] == '/' {
			return i + 2
		}
		i++
	}
	return n
}

// extractJSImports 从源码中提取所有 import 的模块路径
func extractJSImports(content string) []string {
	var imports []string

	// 移除注释
	cleanContent := removeComments(content)
	// 构建字符串掩码，标记哪些位置在字符串内部
	mask := buildStringMask(cleanContent)

	// 使用 FindAllStringSubmatchIndex 减少内存分配
	for _, match := range reESMImport.FindAllStringSubmatchIndex(cleanContent, -1) {
		// match[0] 是整个匹配的起始位置，match[2] 是第一个捕获组的起始位置
		// 如果整个匹配的起始位置在字符串内部，跳过
		if mask[match[0]] {
			continue
		}
		imports = append(imports, cleanContent[match[2]:match[3]])
	}
	for _, match := range reRequire.FindAllStringSubmatchIndex(cleanContent, -1) {
		if mask[match[0]] {
			continue
		}
		imports = append(imports, cleanContent[match[2]:match[3]])
	}
	for _, match := range reDynamicImport.FindAllStringSubmatchIndex(cleanContent, -1) {
		if mask[match[0]] {
			continue
		}
		imports = append(imports, cleanContent[match[2]:match[3]])
	}
	for _, match := range reExportFrom.FindAllStringSubmatchIndex(cleanContent, -1) {
		if mask[match[0]] {
			continue
		}
		imports = append(imports, cleanContent[match[2]:match[3]])
	}

	return imports
}

// resolveJSPackageName 从 import path 中提取包名
// 例如: "lodash/get" -> "lodash", "@org/pkg/sub" -> "@org/pkg"
func resolveJSPackageName(importPath string) string {
	// 跳过相对路径和内置模块
	if len(importPath) == 0 || importPath[0] == '.' || importPath[0] == '/' {
		return ""
	}

	// 查找第一个斜杠
	slashIdx := strings.Index(importPath, "/")
	if slashIdx == -1 {
		return importPath
	}

	// scoped package: @org/pkg
	if importPath[0] == '@' {
		// 查找第二个斜杠
		secondSlash := strings.Index(importPath[slashIdx+1:], "/")
		if secondSlash == -1 {
			return importPath
		}
		return importPath[:slashIdx+1+secondSlash]
	}

	return importPath[:slashIdx]
}
