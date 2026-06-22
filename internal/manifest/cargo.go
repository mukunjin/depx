package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CargoManifest 解析 Cargo.toml 文件
type CargoManifest struct {
	dir     string
	deps    []string
	devDeps []string
}

// NewCargoManifest 创建 Cargo.toml 清单解析器
func NewCargoManifest(dir string) (*CargoManifest, error) {
	cargoPath := filepath.Join(dir, "Cargo.toml")
	if _, err := os.Stat(cargoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cargo.toml not found in %s", dir)
	}

	m := &CargoManifest{dir: dir}
	if err := m.parse(); err != nil {
		return nil, err
	}

	return m, nil
}

// Type 返回包管理器类型
func (m *CargoManifest) Type() string {
	return "cargo"
}

// Dependencies 返回声明的运行时依赖包名列表
func (m *CargoManifest) Dependencies() ([]string, error) {
	return append([]string(nil), m.deps...), nil
}

// DevDependencies 返回 dev-dependencies 列表
func (m *CargoManifest) DevDependencies() ([]string, error) {
	return m.devDeps, nil
}

// parse 解析 Cargo.toml 文件
func (m *CargoManifest) parse() error {
	cargoPath := filepath.Join(m.dir, "Cargo.toml")
	file, err := os.Open(cargoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	m.deps = []string{}
	m.devDeps = []string{}
	scanner := bufio.NewScanner(file)
	inDeps := false
	inDevDeps := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 检测 [dependencies]、[dev-dependencies]、[build-dependencies] 段
		if line == "[dependencies]" {
			inDeps = true
			inDevDeps = false
			continue
		}
		if line == "[dev-dependencies]" {
			inDevDeps = true
			inDeps = false
			continue
		}
		if line == "[build-dependencies]" {
			// treat build-dependencies as dev deps
			inDevDeps = true
			inDeps = false
			continue
		}

		// 检测其他段（结束 dependencies）
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}

		// 在依赖段内提取依赖
		if inDeps || inDevDeps {
			// 解析依赖行：name = "version" 或 name = { version = "x", ... }
			if idx := strings.Index(line, "="); idx > 0 {
				name := strings.TrimSpace(line[:idx])
				if name != "" {
					if inDevDeps {
						m.devDeps = append(m.devDeps, name)
					} else {
						m.deps = append(m.deps, name)
					}
				}
			}
		}
	}

	return scanner.Err()
}
