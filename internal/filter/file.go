package filter

import (
	"path/filepath"
	"strings"
)

// ShouldExcludeFile 检查文件是否应该被排除
func ShouldExcludeFile(path string, patterns []string) bool {
	if path == "" {
		return false
	}

	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}

		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		// 规范化路径分隔符后再比较（Windows 兼容性）
		normalizedPath := filepath.ToSlash(path)
		normalizedPattern := filepath.ToSlash(pattern)
		if strings.Contains(normalizedPath, normalizedPattern) {
			return true
		}
	}
	return false
}

// ShouldExcludeDir 检查目录是否应该被排除
func ShouldExcludeDir(dir string, excludeDirs []string) bool {
	if dir == "" {
		return false
	}

	for _, excludeDir := range excludeDirs {
		if excludeDir == "" {
			continue
		}

		if dir == excludeDir {
			return true
		}
	}
	return false
}
