package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NpmManifest 解析 package.json
type NpmManifest struct {
	path string
	pkg  *npmPackageJSON
}

type npmPackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// NewNpmManifest 创建 npm manifest 解析器
func NewNpmManifest(dir string) (*NpmManifest, error) {
	p := filepath.Join(dir, "package.json")
	if _, err := os.Stat(p); err != nil {
		return nil, fmt.Errorf("package.json not found in %s", dir)
	}
	return &NpmManifest{path: p}, nil
}

func (n *NpmManifest) Type() string { return "npm" }

func (n *NpmManifest) load() (*npmPackageJSON, error) {
	if n.pkg != nil {
		return n.pkg, nil
	}

	data, err := os.ReadFile(n.path)
	if err != nil {
		return nil, err
	}

	var pkg npmPackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", n.path, err)
	}

	n.pkg = &pkg
	return n.pkg, nil
}

func (n *NpmManifest) Dependencies() ([]string, error) {
	pkg, err := n.load()
	if err != nil {
		return nil, err
	}

	var deps []string
	for name := range pkg.Dependencies {
		deps = append(deps, name)
	}
	return deps, nil
}

// DevDependencies 返回 package.json 中的 devDependencies
func (n *NpmManifest) DevDependencies() ([]string, error) {
	pkg, err := n.load()
	if err != nil {
		return nil, err
	}

	var devs []string
	for name := range pkg.DevDependencies {
		devs = append(devs, name)
	}
	return devs, nil
}
