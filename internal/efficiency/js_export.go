package efficiency

import (
	"regexp"
	"strings"
)

// JS/TS import 提取正则
var (
	// import { debounce, throttle } from "lodash"
	reNamedImport = regexp.MustCompile(`import\s*\{([^}]+)\}\s*from\s*['"]([^'"]+)['"]`)
	// import * as _ from "lodash"
	reNamespaceImport = regexp.MustCompile(`import\s*\*\s*as\s+\w+\s+from\s*['"]([^'"]+)['"]`)
	// import lodash from "lodash"
	reDefaultImport = regexp.MustCompile(`import\s+(\w+)\s+from\s*['"]([^'"]+)['"]`)
)

// JS/TS export 提取正则
var (
	// export { foo, bar }
	reExport = regexp.MustCompile(`export\s*\{([^}]+)\}`)
	// export function foo
	reExportFunc = regexp.MustCompile(`export\s+function\s+(\w+)`)
	// export const foo
	reExportConst = regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`)
	// export class Foo
	reExportClass = regexp.MustCompile(`export\s+class\s+(\w+)`)
)

// jsImportInfo 表示一个 JS/TS import 的信息
type jsImportInfo struct {
	Package string   // 包名
	Exports []string // 使用的导出名称
}

// extractJSImportInfos 从 JS/TS 代码中提取所有 import 信息（包名 + 使用的导出）
func extractJSImportInfos(content string) []jsImportInfo {
	var infos []jsImportInfo

	// 分析命名导入：import { foo, bar } from "package"
	for _, match := range reNamedImport.FindAllStringSubmatch(content, -1) {
		if len(match) < 3 {
			continue
		}
		exportList := match[1]
		pkg := match[2]

		var exports []string
		for _, name := range strings.Split(exportList, ",") {
			name = strings.TrimSpace(name)
			// 处理别名：foo as bar（使用 Fields 精确匹配空格分隔）
			if parts := strings.Fields(name); len(parts) >= 3 && parts[1] == "as" {
				name = parts[0]
			}
			if name != "" {
				exports = append(exports, name)
			}
		}
		if len(exports) > 0 {
			infos = append(infos, jsImportInfo{Package: pkg, Exports: exports})
		}
	}

	// 分析命名空间导入：import * as _ from "lodash"
	for _, match := range reNamespaceImport.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		infos = append(infos, jsImportInfo{Package: match[1], Exports: []string{"*"}})
	}

	// 分析默认导入：import lodash from "lodash"
	for _, match := range reDefaultImport.FindAllStringSubmatch(content, -1) {
		if len(match) < 3 {
			continue
		}
		infos = append(infos, jsImportInfo{Package: match[2], Exports: []string{"default"}})
	}

	return infos
}

// extractJSExports 从 JS/TS 源文件内容中提取所有导出名称
func extractJSExports(content string) []string {
	var exports []string

	for _, match := range reExport.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			for _, name := range strings.Split(match[1], ",") {
				name = strings.TrimSpace(name)
				if name != "" {
					exports = append(exports, name)
				}
			}
		}
	}

	for _, match := range reExportFunc.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	for _, match := range reExportConst.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	for _, match := range reExportClass.FindAllStringSubmatch(content, -1) {
		if len(match) >= 2 {
			exports = append(exports, match[1])
		}
	}

	return exports
}
