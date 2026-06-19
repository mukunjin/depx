package usage

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/mukunjin/depx/internal/manifest"
)

// GoAnalyzer 分析 Go 文件中的 import 使用
type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer { return &GoAnalyzer{} }

var goSkipDirs = map[string]bool{
	"vendor":   true,
	".git":     true,
	"testdata": true,
}

func (g *GoAnalyzer) Analyze(dir string, deps []string, opts *Options) (map[string]*manifest.UsageResult, error) {
	// 检查目录是否存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, err
	}

	result := make(map[string]*manifest.UsageResult)
	for _, dep := range deps {
		result[dep] = &manifest.UsageResult{Package: dep}
	}

	// 为每个依赖创建文件追踪 map
	fileTrackers := make(map[string]map[string]bool, len(deps))
	for _, dep := range deps {
		fileTrackers[dep] = make(map[string]bool)
	}

	// 构建跳过目录集合
	skipDirs := make(map[string]bool)
	for k, v := range goSkipDirs {
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

		if strings.ToLower(filepath.Ext(info.Name())) != ".go" {
			return nil
		}

		// 检查文件模式排除
		if opts != nil && shouldExcludeFile(path, opts.ExcludeFiles) {
			return nil
		}

		// 读取文件内容
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		imports := extractGoImports(string(data))

		for _, imp := range imports {
			// 对每个依赖做前缀匹配
			for _, dep := range deps {
				if imp == dep || strings.HasPrefix(imp, dep+"/") {
					r := result[dep]
					r.Used = true
					r.RefCount++

					// 使用 map 追踪文件，避免重复添加
					tracker := fileTrackers[dep]
					if !tracker[relPath] {
						tracker[relPath] = true
						r.UsedIn = append(r.UsedIn, relPath)
					}
				}
			}
		}

		return nil
	})

	return result, err
}

// extractGoImports 从 Go 文件内容中提取 import 路径
func extractGoImports(content string) []string {
	var imports []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	inImport := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行
		if line == "" {
			continue
		}

		// 跳过整行注释
		if strings.HasPrefix(line, "//") {
			continue
		}

		// 移除行尾注释（正确处理字符串字面量）
		line = removeGoLineComment(line)

		// import 块 - 支持 import ( 和 import(
		if strings.HasPrefix(line, "import") && strings.Contains(line, "(") {
			inImport = true
			continue
		}
		if line == ")" {
			inImport = false
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
		if inImport {
			if imp := parseGoImportLine(line); imp != "" {
				imports = append(imports, imp)
			}
		}
	}

	return imports
}

// removeGoLineComment 移除行尾注释，正确处理字符串字面量
// 例如: `"fmt" // comment` -> `"fmt"`
// 例如: `"github.com/foo/bar" // comment` -> `"github.com/foo/bar"`
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
		// 没有结束引号，返回原行
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
// 输入: `"fmt"` 或 `alias "fmt"` 或 `"github.com/foo/bar"`
// 正确处理字符串字面量，不会误删字符串内的 //
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
