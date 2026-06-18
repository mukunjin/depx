package lockfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LockFile 定义 lock file 解析器接口
type LockFile interface {
	// Type 返回包管理器类型
	Type() string
	// Dependencies 返回所有依赖（包括传递依赖）
	Dependencies() ([]Dependency, error)
}

// Dependency 表示 lock file 中的依赖
type Dependency struct {
	Name     string   // 包名
	Version  string   // 版本
	Resolved string   // 完整解析路径
	Requires []string // 依赖的其他包
}

// DetectLockFile 检测并返回合适的 lock file 解析器
func DetectLockFile(dir string) (LockFile, error) {
	// 尝试 npm
	if lf, err := NewNpmLockFile(dir); err == nil {
		return lf, nil
	}

	// 尝试 Go
	if lf, err := NewGoLockFile(dir); err == nil {
		return lf, nil
	}

	// 尝试 Rust
	if lf, err := NewRustLockFile(dir); err == nil {
		return lf, nil
	}

	return nil, fmt.Errorf("no supported lock file found in %s", dir)
}

// NpmLockFile 解析 package-lock.json
type NpmLockFile struct {
	dir  string
	data *npmLockData
}

type npmLockData struct {
	Name            string                    `json:"name"`
	Version         string                    `json:"version"`
	LockfileVersion int                       `json:"lockfileVersion"`
	Dependencies    map[string]npmLockPackage `json:"dependencies"`
	Packages        map[string]npmLockPackage `json:"packages"`
}

type npmLockPackage struct {
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved"`
	Dependencies map[string]string `json:"dependencies"`
	Dev          bool              `json:"dev"`
}

// NewNpmLockFile 创建 npm lock file 解析器
func NewNpmLockFile(dir string) (*NpmLockFile, error) {
	lockPath := filepath.Join(dir, "package-lock.json")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("package-lock.json not found")
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	var lockData npmLockData
	if err := json.Unmarshal(data, &lockData); err != nil {
		return nil, err
	}

	return &NpmLockFile{dir: dir, data: &lockData}, nil
}

// Type 返回包管理器类型
func (lf *NpmLockFile) Type() string {
	return "npm"
}

// Dependencies 返回所有依赖
func (lf *NpmLockFile) Dependencies() ([]Dependency, error) {
	var deps []Dependency

	// 优先使用 packages（lockfileVersion 2+）
	if lf.data.Packages != nil {
		for name, pkg := range lf.data.Packages {
			// 跳过根包
			if name == "" {
				continue
			}

			// 移除 "node_modules/" 前缀
			cleanName := strings.TrimPrefix(name, "node_modules/")

			deps = append(deps, Dependency{
				Name:     cleanName,
				Version:  pkg.Version,
				Resolved: pkg.Resolved,
			})
		}
	} else if lf.data.Dependencies != nil {
		// 使用旧版 dependencies 字段
		for name, pkg := range lf.data.Dependencies {
			var requires []string
			for req := range pkg.Dependencies {
				requires = append(requires, req)
			}

			deps = append(deps, Dependency{
				Name:     name,
				Version:  pkg.Version,
				Resolved: pkg.Resolved,
				Requires: requires,
			})
		}
	}

	return deps, nil
}

// GoLockFile 解析 go.sum
type GoLockFile struct {
	dir   string
	lines []string
}

// NewGoLockFile 创建 Go lock file 解析器
func NewGoLockFile(dir string) (*GoLockFile, error) {
	lockPath := filepath.Join(dir, "go.sum")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("go.sum not found")
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	return &GoLockFile{dir: dir, lines: lines}, nil
}

// Type 返回包管理器类型
func (lf *GoLockFile) Type() string {
	return "go"
}

// Dependencies 返回所有依赖
func (lf *GoLockFile) Dependencies() ([]Dependency, error) {
	var deps []Dependency
	seen := make(map[string]bool)

	for _, line := range lf.lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// go.sum 格式: module/path version hash
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		name := parts[0]
		version := parts[1]

		// 去重（go.sum 中同一模块可能出现多次）
		key := name + "@" + version
		if seen[key] {
			continue
		}
		seen[key] = true

		deps = append(deps, Dependency{
			Name:    name,
			Version: version,
		})
	}

	return deps, nil
}

// RustLockFile 解析 Cargo.lock
type RustLockFile struct {
	dir     string
	content string
}

// NewRustLockFile 创建 Rust lock file 解析器
func NewRustLockFile(dir string) (*RustLockFile, error) {
	lockPath := filepath.Join(dir, "Cargo.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cargo.lock not found")
	}

	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	return &RustLockFile{dir: dir, content: string(data)}, nil
}

// Type 返回包管理器类型
func (lf *RustLockFile) Type() string {
	return "cargo"
}

// Dependencies 返回所有依赖
func (lf *RustLockFile) Dependencies() ([]Dependency, error) {
	var deps []Dependency

	// 简单解析 TOML 格式的 Cargo.lock
	// 查找 [[package]] 块
	blocks := strings.Split(lf.content, "[[package]]")

	for _, block := range blocks[1:] { // 跳过第一个空块
		lines := strings.Split(block, "\n")
		var name, version string

		for _, line := range lines {
			line = strings.TrimSpace(line)

			if strings.HasPrefix(line, "name = ") {
				name = strings.Trim(strings.TrimPrefix(line, "name = "), "\"")
			} else if strings.HasPrefix(line, "version = ") {
				version = strings.Trim(strings.TrimPrefix(line, "version = "), "\"")
			}

			if name != "" && version != "" {
				deps = append(deps, Dependency{
					Name:    name,
					Version: version,
				})
				break
			}
		}
	}

	return deps, nil
}
