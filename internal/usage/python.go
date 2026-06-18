package usage

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mukunjin/depx/internal/manifest"
)

// PythonAnalyzer 分析 Python 文件中的 import 语句
type PythonAnalyzer struct{}

// NewPythonAnalyzer 创建 Python 分析器
func NewPythonAnalyzer() *PythonAnalyzer {
	return &PythonAnalyzer{}
}

// 匹配的源文件扩展名
var pythonExtensions = map[string]bool{
	".py": true,
}

// 跳过的目录
var pythonSkipDirs = map[string]bool{
	"__pycache__":  true,
	".git":         true,
	"venv":         true,
	".venv":        true,
	"env":          true,
	"node_modules": true,
}

// import 匹配正则
var (
	// import requests
	rePythonImport = regexp.MustCompile(`(?m)^\s*import\s+([a-zA-Z_][a-zA-Z0-9_.]*)`)
	// from flask import Flask
	rePythonFromImport = regexp.MustCompile(`(?m)^\s*from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+import`)
)

// Analyze 扫描目录下所有 Python 源文件，返回每个依赖的使用情况
func (p *PythonAnalyzer) Analyze(dir string, deps []string) (map[string]*manifest.UsageResult, error) {
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
			if pythonSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(info.Name())
		if !pythonExtensions[ext] {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		relPath, _ := filepath.Rel(dir, path)

		// 提取所有 import 的包名
		imports := extractPythonImports(content)
		for _, imp := range imports {
			pkgName := resolvePythonPackageName(imp)
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

// extractPythonImports 从源码中提取所有 import 的模块路径
func extractPythonImports(content string) []string {
	var imports []string

	// 移除注释
	cleanContent := removePythonComments(content)

	// 按行处理，保持源代码顺序
	lines := strings.Split(cleanContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 尝试匹配 import 语句
		if match := rePythonImport.FindStringSubmatchIndex(line); match != nil {
			imports = append(imports, line[match[2]:match[3]])
			continue
		}

		// 尝试匹配 from...import 语句
		if match := rePythonFromImport.FindStringSubmatchIndex(line); match != nil {
			imports = append(imports, line[match[2]:match[3]])
		}
	}

	return imports
}

// removePythonComments 移除 Python 代码中的注释
func removePythonComments(content string) string {
	var result strings.Builder
	result.Grow(len(content))

	i, n := 0, len(content)
	for i < n {
		c := content[i]

		// 处理字符串字面量
		if c == '"' || c == '\'' {
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
		if c == '#' {
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

		result.WriteByte(c)
		i++
	}

	return result.String()
}

// resolvePythonPackageName 从 import path 中提取包名
// 例如: "flask" -> "flask", "requests.auth" -> "requests"
func resolvePythonPackageName(importPath string) string {
	if len(importPath) == 0 {
		return ""
	}

	// 查找第一个点
	idx := strings.Index(importPath, ".")
	if idx == -1 {
		return importPath
	}

	return importPath[:idx]
}
