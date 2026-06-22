package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GoModManifest 解析 go.mod
type GoModManifest struct {
	path string
}

// NewGoModManifest 创建 go.mod 解析器
func NewGoModManifest(dir string) (*GoModManifest, error) {
	p := filepath.Join(dir, "go.mod")
	if _, err := os.Stat(p); err != nil {
		return nil, fmt.Errorf("go.mod not found in %s", dir)
	}
	return &GoModManifest{path: p}, nil
}

func (g *GoModManifest) Type() string { return "go" }

func (g *GoModManifest) Dependencies() ([]string, error) {
	f, err := os.Open(g.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var deps []string
	inRequire := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过注释和空行
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// 处理 require 块
		if strings.HasPrefix(line, "require(") || strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}

		// 单行 require
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				modPath := parts[1]
				if isDirectDep(line) && isThirdParty(modPath) {
					deps = append(deps, modPath)
				}
			}
			continue
		}

		// require 块内的行
		if inRequire {
			// 跳过 indirect 依赖
			if strings.Contains(line, "// indirect") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				modPath := parts[0]
				if isThirdParty(modPath) {
					deps = append(deps, modPath)
				}
			}
		}
	}

	return deps, scanner.Err()
}

// DevDependencies 对于 go.mod 返回空（Go 没有单独的 devDependencies）
func (g *GoModManifest) DevDependencies() ([]string, error) {
	return []string{}, nil
}

// isDirectDep 检查是否是直接依赖（非 indirect）
func isDirectDep(line string) bool {
	return !strings.Contains(line, "// indirect")
}

// isThirdParty 检查是否是第三方包（标准库不含 '.'）
func isThirdParty(modPath string) bool {
	return strings.Contains(modPath, ".")
}
