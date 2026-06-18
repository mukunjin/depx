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

func (n *NpmManifest) Dependencies() ([]string, error) {
	data, err := os.ReadFile(n.path)
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", n.path, err)
	}

	seen := make(map[string]bool)
	var deps []string
	for name := range pkg.Dependencies {
		if !seen[name] {
			seen[name] = true
			deps = append(deps, name)
		}
	}
	for name := range pkg.DevDependencies {
		if !seen[name] {
			seen[name] = true
			deps = append(deps, name)
		}
	}
	return deps, nil
}
