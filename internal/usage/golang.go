package usage

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/depx/depx/internal/manifest"
)

// GoAnalyzer 分析 Go 文件中的 import 使用
type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer { return &GoAnalyzer{} }

var goSkipDirs = map[string]bool{
	"vendor":   true,
	".git":     true,
	"testdata": true,
}

func (g *GoAnalyzer) Analyze(dir string, deps []string) (map[string]*manifest.UsageResult, error) {
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

	// 构建依赖集合用于快速查找
	depSet := make(map[string]bool, len(deps))
	for _, dep := range deps {
		depSet[dep] = true
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if goSkipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(info.Name()) != ".go" {
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

		// 跳过单行注释（整行都是注释）
		if strings.HasPrefix(line, "//") {
			continue
		}

		// 移除行尾注释
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

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

// parseGoImportLine 解析单行 import，支持别名
// 输入: `"fmt"` 或 `alias "fmt"` 或 `"github.com/foo/bar"`
func parseGoImportLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// 去掉可能的行尾注释
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}

	// 找到引号内的内容
	start := strings.Index(line, `"`)
	end := strings.LastIndex(line, `"`)
	if start < 0 || end <= start {
		return ""
	}

	return line[start+1 : end]
}
