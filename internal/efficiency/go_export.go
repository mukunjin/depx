package efficiency

import (
	"regexp"
	"strings"
)

// Go 导出提取正则（大写字母开头的标识符为导出）
var (
	reGoExportFunc  = regexp.MustCompile(`(?m)^func\s+([A-Z][a-zA-Z0-9_]*)\s*\(`)
	reGoExportType  = regexp.MustCompile(`(?m)^type\s+([A-Z][a-zA-Z0-9_]*)\s+(struct|interface)`)
	reGoExportConst = regexp.MustCompile(`(?m)^const\s+([A-Z][a-zA-Z0-9_]*)`)
	reGoExportVar   = regexp.MustCompile(`(?m)^var\s+([A-Z][a-zA-Z0-9_]*)`)
)

// extractGoExports 从 Go 源文件内容中提取所有导出名称
func extractGoExports(content string) []string {
	var exports []string

	for _, match := range reGoExportFunc.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	for _, match := range reGoExportType.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	for _, match := range reGoExportConst.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	for _, match := range reGoExportVar.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	return exports
}

// extractGoImportPaths 从 Go 代码中提取所有 import 路径
func extractGoImportPaths(content string) []string {
	var imports []string
	lines := strings.Split(content, "\n")
	inImportBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "import (") {
			inImportBlock = true
			continue
		}
		if line == ")" {
			inImportBlock = false
			continue
		}

		if inImportBlock {
			if imp := parseGoImportPath(line); imp != "" {
				imports = append(imports, imp)
			}
			continue
		}

		if strings.HasPrefix(line, "import ") && !strings.Contains(line, "(") {
			if imp := parseGoImportPath(line[7:]); imp != "" {
				imports = append(imports, imp)
			}
		}
	}

	return imports
}

// parseGoImportPath 从一行 import 语句中提取路径
func parseGoImportPath(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "//") {
		return ""
	}
	start := strings.Index(line, `"`)
	if start < 0 {
		return ""
	}
	end := strings.Index(line[start+1:], `"`)
	if end < 0 {
		return ""
	}
	return line[start+1 : start+1+end]
}
