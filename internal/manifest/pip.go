package manifest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PipManifest 解析 requirements.txt 文件
type PipManifest struct {
	dir  string
	deps []string
}

// NewPipManifest 创建 requirements.txt 清单解析器
func NewPipManifest(dir string) (*PipManifest, error) {
	reqPath := filepath.Join(dir, "requirements.txt")
	if _, err := os.Stat(reqPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("requirements.txt not found in %s", dir)
	}

	m := &PipManifest{dir: dir}
	if err := m.parse(); err != nil {
		return nil, err
	}

	return m, nil
}

// Type 返回包管理器类型
func (m *PipManifest) Type() string {
	return "pip"
}

// Dependencies 返回声明的依赖包名列表
func (m *PipManifest) Dependencies() ([]string, error) {
	return m.deps, nil
}

// parse 解析 requirements.txt 文件
func (m *PipManifest) parse() error {
	reqPath := filepath.Join(m.dir, "requirements.txt")
	file, err := os.Open(reqPath)
	if err != nil {
		return err
	}
	defer file.Close()

	m.deps = []string{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 跳过 -r 和 -e 选项
		if strings.HasPrefix(line, "-r") || strings.HasPrefix(line, "-e") {
			continue
		}

		// 提取包名（去除版本约束）
		pkgName := extractPipPackageName(line)
		if pkgName != "" {
			m.deps = append(m.deps, pkgName)
		}
	}

	return scanner.Err()
}

// extractPipPackageName 从依赖行中提取包名
// 例如: "requests==2.28.0" -> "requests", "numpy>=1.21" -> "numpy"
func extractPipPackageName(line string) string {
	// 查找版本约束符号
	for _, sep := range []string{"==", ">=", "<=", "~=", "!=", ">", "<"} {
		if idx := strings.Index(line, sep); idx > 0 {
			return strings.TrimSpace(line[:idx])
		}
	}

	// 没有版本约束，直接返回
	return strings.TrimSpace(line)
}
