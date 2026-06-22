package surface

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/mukunjin/depx/internal/filter"
)

// SurfaceResult 影响面分析结果
type SurfaceResult struct {
	Package     string              // 包名
	Files       []string            // 使用该包的文件列表
	Modules     []string            // 使用的子模块/子路径列表
	RefCount    int                 // 总引用次数
	Score       int                 // 影响面分数 (RefCount*5 + Modules)
	Criticality string              // 关键度：High/Medium/Low
	fileSet     map[string]struct{} // 用于文件去重
	moduleSet   map[string]struct{} // 用于模块去重
}

// CalculateScore 计算影响面分数
func (r *SurfaceResult) CalculateScore() {
	r.Score = r.RefCount*5 + len(r.Modules)
}

// 百分位关键度阈值
const (
	highCriticalityThreshold   = 0.2 // Top 20% = High
	mediumCriticalityThreshold = 0.5 // Top 50% = Medium
)

// 语言类型
type langType int

const (
	langJS langType = iota
	langGo
	langRust
	langPython
)

// typesPrefix 是 TypeScript 类型包的统一前缀
const typesPrefix = "@types/"

// 正则表达式匹配 import 语句
var (
	// JS/TS
	reImport        = regexp.MustCompile(`import\s+(?:[\w*{}\s,]+?\s+from\s+)?['"]([^'"]+)['"]`)
	reRequire       = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	reDynamicImport = regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	reExportFrom    = regexp.MustCompile(`export\s+(?:[\w*{}\s,]+?\s+from\s+)['"]([^'"]+)['"]`)
)

// 支持的文件扩展名
var supportedExts = map[string]langType{
	".js":     langJS,
	".ts":     langJS,
	".jsx":    langJS,
	".tsx":    langJS,
	".mjs":    langJS,
	".cjs":    langJS,
	".vue":    langJS,
	".svelte": langJS,
	".go":     langGo,
	".rs":     langRust,
	".py":     langPython,
}

// Options 定义影响面分析选项
type Options struct {
	// ManifestType 项目类型 ("npm", "go", "cargo", "pip")
	ManifestType string
	// ExcludeDirs 排除的目录
	ExcludeDirs []string
	// ExcludeFiles 排除的文件模式
	ExcludeFiles []string
	// ReadNodeModules 是否读取 node_modules
	ReadNodeModules bool
}

var manifestExts = map[string]map[string]langType{
	"npm": {
		".js": langJS, ".ts": langJS, ".jsx": langJS, ".tsx": langJS,
		".mjs": langJS, ".cjs": langJS, ".vue": langJS, ".svelte": langJS,
	},
	"go":    {".go": langGo},
	"cargo": {".rs": langRust},
	"pip":   {".py": langPython},
}

var defaultSkipDirs = map[string][]string{
	"npm":   {"node_modules", "dist", "build", ".next", ".nuxt", "coverage"},
	"go":    {"vendor"},
	"cargo": {"target"},
	"pip":   {"__pycache__", ".venv", "venv"},
}

// AnalyzeSurface 分析依赖的影响面
func AnalyzeSurface(dir string, deps []string, opts *Options) (map[string]*SurfaceResult, error) {
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

	// 按长度降序排序依赖列表，优先匹配最长的
	sortedDeps := make([]string, len(deps))
	copy(sortedDeps, deps)
	sort.Slice(sortedDeps, func(i, j int) bool {
		return len(sortedDeps[i]) > len(sortedDeps[j])
	})

	// 构建跳过目录集合
	skipDirs := make(map[string]bool)
	if opts != nil {
		if !opts.ReadNodeModules {
			skipDirs["node_modules"] = true
		}
		for _, d := range defaultSkipDirs[opts.ManifestType] {
			skipDirs[d] = true
		}
		for _, d := range opts.ExcludeDirs {
			skipDirs[d] = true
		}
	}

	allowedExts := supportedExts
	if opts != nil && opts.ManifestType != "" {
		if exts, ok := manifestExts[opts.ManifestType]; ok {
			allowedExts = exts
		}
	}

	// 扫描所有源文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过指定目录
		if info.IsDir() {
			name := info.Name()
			if skipDirs[name] || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件模式排除
		if opts != nil && filter.ShouldExcludeFile(path, opts.ExcludeFiles) {
			return nil
		}

		// 检查文件扩展名
		ext := strings.ToLower(filepath.Ext(info.Name()))
		lang, supported := allowedExts[ext]
		if !supported {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(data)
		relPath, _ := filepath.Rel(dir, path)

		// 提取所有 import
		imports := extractImportsByLang(content, lang)
		for _, imp := range imports {
			// 使用最长前缀匹配
			pkgName := matchDependency(imp, sortedDeps, lang)
			if pkgName != "" {
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

	// 计算分数并收集所有结果
	var allResults []*SurfaceResult
	for _, result := range results {
		result.CalculateScore()
		allResults = append(allResults, result)
	}

	// 按分数排序用于计算百分位
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	// 根据百分位计算关键度
	totalCount := len(allResults)
	for i, result := range allResults {
		percentile := float64(i) / float64(totalCount)
		if percentile < highCriticalityThreshold {
			result.Criticality = "High"
		} else if percentile < mediumCriticalityThreshold {
			result.Criticality = "Medium"
		} else {
			result.Criticality = "Low"
		}

		sort.Strings(result.Files)
		sort.Strings(result.Modules)
	}

	return results, nil
}

// matchDependency 使用最长前缀匹配查找依赖
func matchDependency(importPath string, sortedDeps []string, lang langType) string {
	for _, dep := range sortedDeps {
		if isImportOfDep(importPath, dep, lang) {
			return dep
		}
	}
	return ""
}

// isImportOfDep 检查 import 路径是否属于某个依赖
func isImportOfDep(importPath, dep string, lang langType) bool {
	// 精确匹配
	if importPath == dep {
		return true
	}

	// 根据语言判断分隔符
	var sep string
	switch lang {
	case langGo:
		sep = "/"
	case langPython:
		sep = "."
	case langJS:
		sep = "/"
	case langRust:
		sep = "::"
	}

	// 前缀匹配（import 路径以 dep + 分隔符开头）
	return strings.HasPrefix(importPath, dep+sep)
}

// extractImportsByLang 根据语言提取 import
func extractImportsByLang(content string, lang langType) []string {
	switch lang {
	case langGo:
		return extractGoImports(content)
	case langRust:
		return extractRustImports(content)
	case langPython:
		return extractPythonImports(content)
	default:
		return extractJSImports(content)
	}
}

// extractJSImports 提取 JS/TS import
func extractJSImports(content string) []string {
	var imports []string

	// 移除注释并构建字符串掩码
	cleanContent := removeJSComments(content)
	mask := buildJSStringMask(cleanContent)

	// 使用 FindAllStringSubmatchIndex 减少内存分配
	for _, match := range reImport.FindAllStringSubmatchIndex(cleanContent, -1) {
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

// removeJSComments 移除 JS/TS 代码中的注释
func removeJSComments(content string) string {
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
				// 单行注释
				for i < n && content[i] != '\n' {
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

// buildJSStringMask 构建字符串区域掩码
func buildJSStringMask(content string) []bool {
	mask := make([]bool, len(content))
	i, n := 0, len(content)

	for i < n {
		c := content[i]

		if c == '"' || c == '\'' || c == '`' {
			quote := c
			mask[i] = true
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

// extractGoImports 提取 Go import
func extractGoImports(content string) []string {
	var imports []string

	// 使用状态机解析
	lines := strings.Split(content, "\n")
	inImportBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// 移除行尾注释
		line = removeGoLineComment(line)

		// import 块开始
		if strings.HasPrefix(line, "import") && strings.Contains(line, "(") {
			inImportBlock = true
			continue
		}

		// import 块结束
		if inImportBlock && line == ")" {
			inImportBlock = false
			continue
		}

		// 单行 import
		if strings.HasPrefix(line, "import ") && !strings.Contains(line, "(") {
			if imp := parseGoImportLine(line[7:]); imp != "" {
				imports = append(imports, imp)
			}
			continue
		}

		// import 块内的行
		if inImportBlock {
			if imp := parseGoImportLine(line); imp != "" {
				imports = append(imports, imp)
			}
		}
	}

	return imports
}

// removeGoLineComment 移除行尾注释
func removeGoLineComment(line string) string {
	// 查找第一个引号
	quoteIdx := strings.Index(line, `"`)
	if quoteIdx < 0 {
		// 没有字符串字面量，直接查找注释
		if idx := strings.Index(line, "//"); idx >= 0 {
			return strings.TrimSpace(line[:idx])
		}
		return line
	}

	// 查找字符串字面量的结束引号
	endQuoteIdx := strings.Index(line[quoteIdx+1:], `"`)
	if endQuoteIdx < 0 {
		return line
	}

	// 计算结束引号的实际位置
	endQuoteIdx += quoteIdx + 1

	// 在字符串字面量之后查找注释
	afterQuote := line[endQuoteIdx+1:]
	if idx := strings.Index(afterQuote, "//"); idx >= 0 {
		return strings.TrimSpace(line[:endQuoteIdx+1+idx])
	}

	return line
}

// parseGoImportLine 解析单行 import，支持别名
func parseGoImportLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// 找到第一个引号
	start := strings.Index(line, `"`)
	if start < 0 {
		return ""
	}

	// 从引号开始查找结束引号
	end := strings.Index(line[start+1:], `"`)
	if end < 0 {
		return ""
	}

	// 提取引号内的内容
	return line[start+1 : start+1+end]
}

// extractRustImports 提取 Rust use/extern crate
func extractRustImports(content string) []string {
	var imports []string

	// 移除注释
	cleanContent := removeRustComments(content)

	// 使用状态机解析
	lines := strings.Split(cleanContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行
		if line == "" {
			continue
		}

		// use 语句
		if strings.HasPrefix(line, "use ") {
			// 移除末尾的 ;
			line = strings.TrimSuffix(line, ";")
			line = strings.TrimSpace(line[4:])

			// 处理 use xxx::yyy::{aaa, bbb}
			if strings.Contains(line, "{") {
				// 提取包名（{ 之前的部分）
				parts := strings.Split(line, "{")
				if len(parts) > 0 {
					pkgPath := strings.TrimSpace(parts[0])
					pkgPath = strings.TrimSuffix(pkgPath, "::")
					pkgParts := strings.Split(pkgPath, "::")
					if len(pkgParts) > 0 {
						imports = append(imports, pkgParts[0])
					}
				}
			} else {
				// 普通 use 语句，取第一个路径段
				parts := strings.Split(line, "::")
				if len(parts) > 0 {
					imports = append(imports, parts[0])
				}
			}
			continue
		}

		// extern crate 语句
		if strings.HasPrefix(line, "extern crate ") {
			line = strings.TrimSuffix(line, ";")
			line = strings.TrimSpace(line[13:])
			// 处理 extern crate xxx as yyy
			if strings.Contains(line, " as ") {
				parts := strings.Split(line, " as ")
				line = strings.TrimSpace(parts[0])
			}
			if line != "" {
				imports = append(imports, line)
			}
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

// extractPythonImports 提取 Python import
func extractPythonImports(content string) []string {
	var imports []string

	// 移除注释
	cleanContent := removePythonComments(content)

	// 处理多行导入（括号内的导入）
	cleanContent = strings.ReplaceAll(cleanContent, "\\\n", " ")
	cleanContent = strings.ReplaceAll(cleanContent, "(\n", "(")
	cleanContent = strings.ReplaceAll(cleanContent, "\n)", ")")

	// 使用状态机解析
	lines := strings.Split(cleanContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// import xxx 或 import xxx.yyy
		if strings.HasPrefix(line, "import ") {
			line = strings.TrimSpace(line[7:])
			// 处理 import xxx as yyy
			if strings.Contains(line, " as ") {
				parts := strings.Split(line, " as ")
				line = strings.TrimSpace(parts[0])
			}
			// 处理 import xxx, yyy
			for _, imp := range strings.Split(line, ",") {
				imp = strings.TrimSpace(imp)
				if imp != "" && !strings.HasPrefix(imp, ".") {
					imports = append(imports, resolvePythonPackageName(imp))
				}
			}
			continue
		}

		// from xxx import yyy 或 from xxx.yyy import zzz
		if strings.HasPrefix(line, "from ") {
			parts := strings.Split(line, " import ")
			if len(parts) >= 2 {
				pkgPath := strings.TrimSpace(parts[0])
				pkgPath = strings.TrimPrefix(pkgPath, "from ")
				// 跳过相对导入
				if !strings.HasPrefix(pkgPath, ".") {
					imports = append(imports, resolvePythonPackageName(pkgPath))
				}
			}
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

		// 处理三引号字符串 (""" 或 ''')
		if (c == '"' || c == '\'') && i+2 < n && content[i+1] == c && content[i+2] == c {
			// 跳过三引号
			i += 3
			// 找到结束的三引号
			for i+2 < n {
				if content[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				if content[i] == c && content[i+1] == c && content[i+2] == c {
					i += 3
					break
				}
				i++
			}
			// 如果没找到结束，跳过剩余内容
			if i >= n {
				break
			}
			result.WriteString(" None ") // 用占位符替换字符串
			continue
		}

		// 处理单引号字符串字面量
		if c == '"' || c == '\'' {
			quote := c
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
				if content[i] == quote {
					i++
					break
				}
				i++
			}
			continue
		}

		// 处理注释
		if c == '#' {
			// 单行注释
			for i < n && content[i] != '\n' {
				i++
			}
			continue
		}

		result.WriteByte(c)
		i++
	}

	return result.String()
}

// resolvePythonPackageName 解析 Python 包名（取第一段）
func resolvePythonPackageName(importPath string) string {
	parts := strings.Split(importPath, ".")
	return parts[0]
}
