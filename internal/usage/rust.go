package usage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mukunjin/depx/internal/manifest"
)

// RustAnalyzer 分析 Rust 文件中的 use 和 extern crate 语句
type RustAnalyzer struct{}

// NewRustAnalyzer 创建 Rust 分析器
func NewRustAnalyzer() *RustAnalyzer {
	return &RustAnalyzer{}
}

// 匹配的源文件扩展名
var rustExtensions = map[string]bool{
	".rs": true,
}

// 跳过的目录
var rustSkipDirs = map[string]bool{
	"target":       true,
	".git":         true,
	"vendor":       true,
	"node_modules": true,
}

// import 匹配正则
var (
	// use serde::Deserialize;
	reRustUse = regexp.MustCompile(`(?m)^\s*use\s+([a-zA-Z_][a-zA-Z0-9_:]*)`)
	// extern crate tokio;
	reRustExternCrate = regexp.MustCompile(`(?m)^\s*extern\s+crate\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
)

// Analyze 扫描目录下所有 Rust 源文件，返回每个依赖的使用情况
func (r *RustAnalyzer) Analyze(dir string, deps []string, opts *Options) (map[string]*manifest.UsageResult, error) {
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
	for k, v := range rustSkipDirs {
		skipDirs[k] = v
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
		if !rustExtensions[ext] {
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
		imports := extractRustImports(content)
		for _, imp := range imports {
			pkgName := resolveRustPackageName(imp)
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

// extractRustImports 从源码中提取所有 import 的模块路径
func extractRustImports(content string) []string {
	var imports []string

	// 移除注释
	cleanContent := removeRustComments(content)

	// 按行处理，保持源代码顺序
	lines := strings.Split(cleanContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 尝试匹配 use 语句
		if match := reRustUse.FindStringSubmatchIndex(line); match != nil {
			imports = append(imports, line[match[2]:match[3]])
			continue
		}

		// 尝试匹配 extern crate 语句
		if match := reRustExternCrate.FindStringSubmatchIndex(line); match != nil {
			imports = append(imports, line[match[2]:match[3]])
		}
	}

	return imports
}

// removeRustComments 移除 Rust 代码中的注释
func removeRustComments(content string) string {
	var result strings.Builder
	result.Grow(len(content))

	i, n := 0, len(content)
	for i < n {
		c := content[i]

		// 处理字符串字面量
		if c == '"' {
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
				if content[i] == '"' {
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
				// 单行注释
				for i < n && content[i] != '\n' {
					i++
				}
				// 保留换行符
				result.WriteByte('\n')
				if i < n {
					i++
				}
				continue
			}
			if content[i+1] == '*' {
				// 块注释
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

		result.WriteByte(c)
		i++
	}

	return result.String()
}

// resolveRustPackageName 从 import path 中提取包名
// 例如: "serde::Deserialize" -> "serde", "tokio::sync::Mutex" -> "tokio"
func resolveRustPackageName(importPath string) string {
	if len(importPath) == 0 {
		return ""
	}

	// 查找第一个 ::
	idx := strings.Index(importPath, "::")
	if idx == -1 {
		return importPath
	}

	return importPath[:idx]
}
