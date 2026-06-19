package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunSurface(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(string) error
		path        string
		expectError bool
	}{
		{
			name: "surface analysis",
			setup: func(dir string) error {
				// 创建 package.json
				pkgContent := `{"name": "test", "dependencies": {"lodash": "^4.17.21", "axios": "^1.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				// 创建源文件
				srcContent := `import { debounce } from "lodash";
import axios from "axios";
`
				return os.WriteFile(filepath.Join(dir, "index.js"), []byte(srcContent), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name: "surface with path",
			setup: func(dir string) error {
				pkgContent := `{"name": "test", "dependencies": {"react": "^18.0.0"}}`
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgContent), 0644); err != nil {
					return err
				}
				srcContent := `import React from "react";`
				return os.WriteFile(filepath.Join(dir, "app.js"), []byte(srcContent), 0644)
			},
			path:        ".",
			expectError: false,
		},
		{
			name:        "surface non-existent path",
			path:        "/non/existent/path",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			origDir, _ := os.Getwd()
			defer os.Chdir(origDir)

			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			if tt.setup != nil {
				if err := tt.setup(tmpDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err := runSurface(tt.path)
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
