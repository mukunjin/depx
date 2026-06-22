package cmd

import (
	"fmt"

	"github.com/mukunjin/depx/internal/config"
)

func loadConfig(projectPath, configPath string) (*config.Config, error) {
	var (
		cfg *config.Config
		err error
	)

	if configPath != "" {
		cfg, err = config.Load(configPath)
		if err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	} else {
		cfg, err = config.FindAndLoad(projectPath)
		if err != nil {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
