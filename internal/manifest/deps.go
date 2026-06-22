package manifest

// MergeWithDev 合并 runtime 与 dev 依赖（去重）
func MergeWithDev(m Manifest) ([]string, error) {
	deps, err := m.Dependencies()
	if err != nil {
		return nil, err
	}

	devDeps, err := m.DevDependencies()
	if err != nil || len(devDeps) == 0 {
		return deps, nil
	}

	seen := make(map[string]struct{}, len(deps)+len(devDeps))
	for _, dep := range deps {
		seen[dep] = struct{}{}
	}
	for _, dep := range devDeps {
		if _, ok := seen[dep]; !ok {
			deps = append(deps, dep)
			seen[dep] = struct{}{}
		}
	}
	return deps, nil
}
