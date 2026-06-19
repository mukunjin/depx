package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 定义 depx 配置文件结构
type Config struct {
	// Ignore 忽略的依赖（不计入未使用）
	Ignore []string `yaml:"ignore"`

	// ExcludeDirs 忽略的目录
	ExcludeDirs []string `yaml:"exclude_dirs"`

	// ExcludeFiles 忽略的文件模式
	ExcludeFiles []string `yaml:"exclude_files"`

	// ReadNodeModules 是否读取 node_modules 进行精确分析
	ReadNodeModules bool `yaml:"read_node_modules"`

	// LockFile 是否启用 Lock File 分析
	LockFile bool `yaml:"lock_file"`
}

// 默认排除的目录列表
var defaultExcludeDirs = []string{"node_modules", "vendor", "dist", "build"}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Ignore:          []string{},
		ExcludeDirs:     append([]string(nil), defaultExcludeDirs...),
		ExcludeFiles:    []string{},
		ReadNodeModules: false,
		LockFile:        true,
	}
}

// Load 从指定路径加载配置文件
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// FindAndLoad 在项目目录中查找并加载 .depx.yml
func FindAndLoad(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ".depx.yml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认配置
			return DefaultConfig(), nil
		}
		// 其他错误（如权限错误）应该返回
		return nil, err
	}

	return Load(configPath)
}

// IsIgnored 检查依赖是否被忽略
func (c *Config) IsIgnored(pkg string) bool {
	for _, ignored := range c.Ignore {
		if ignored == pkg {
			return true
		}
	}
	return false
}

// IsDirExcluded 检查目录是否被排除
func (c *Config) IsDirExcluded(dir string) bool {
	for _, excluded := range c.ExcludeDirs {
		if excluded == dir {
			return true
		}
	}
	return false
}

// Validate 验证配置的有效性，返回错误如果配置无效
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	// 验证 ExcludeDirs 不包含空字符串
	for _, dir := range c.ExcludeDirs {
		if dir == "" {
			return fmt.Errorf("exclude_dirs contains empty string")
		}
	}

	// 验证 ExcludeFiles 不包含空字符串
	for _, pattern := range c.ExcludeFiles {
		if pattern == "" {
			return fmt.Errorf("exclude_files contains empty string")
		}
	}

	return nil
}
