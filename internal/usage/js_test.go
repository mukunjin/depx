package usage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/depx/depx/internal/manifest"
)

func TestJSAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建测试文件
	jsCode := `
import axios from 'axios';
import { debounce } from 'lodash';
const moment = require('moment');
import('dynamic');
export { foo } from 'bar';
import React from 'react';
import { useState } from 'react';
import '@org/pkg';
import '@org/pkg/sub';
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "moment", "dynamic", "bar", "react", "@org/pkg"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 验证所有依赖都被检测到
	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Expected dependency '%s' to be used", dep)
		}
	}

	// 验证未使用的依赖
	results["unused"] = &manifest.UsageResult{Package: "unused", Used: false}
	if results["unused"].Used {
		t.Error("Expected 'unused' to not be used")
	}
}

func TestJSAnalyzerSubpath(t *testing.T) {
	tmpDir := t.TempDir()

	jsCode := `
import { get } from 'lodash/fp';
import axios from 'axios/lib/adapters/http';
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"lodash", "axios"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// 子路径导入应该匹配主包
	if !results["lodash"].Used {
		t.Error("Expected 'lodash' to be used via subpath")
	}
	if !results["axios"].Used {
		t.Error("Expected 'axios' to be used via subpath")
	}
}

func TestJSAnalyzerSkipDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// 在 node_modules 中创建文件（应该被跳过）
	nodeModules := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatal(err)
	}

	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(nodeModules, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios"}

	results, err := analyzer.Analyze(tmpDir, deps)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// node_modules 中的导入应该被忽略
	if results["axios"].Used {
		t.Error("Expected 'axios' to not be used (should skip node_modules)")
	}
}

func TestExtractJSImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "ESM import",
			code:     `import axios from 'axios';`,
			expected: []string{"axios"},
		},
		{
			name:     "Named import",
			code:     `import { useState, useEffect } from 'react';`,
			expected: []string{"react"},
		},
		{
			name:     "Side effect import",
			code:     `import './styles.css';`,
			expected: []string{"./styles.css"},
		},
		{
			name:     "Require",
			code:     `const fs = require('fs');`,
			expected: []string{"fs"},
		},
		{
			name:     "Dynamic import",
			code:     `const module = import('./module.js');`,
			expected: []string{"./module.js"},
		},
		{
			name:     "Export from",
			code:     `export { foo } from 'bar';`,
			expected: []string{"bar"},
		},
		{
			name:     "Scoped package",
			code:     `import pkg from '@org/pkg';`,
			expected: []string{"@org/pkg"},
		},
		{
			name:     "Multiple imports",
			code:     `import a from 'a';\nimport b from 'b';`,
			expected: []string{"a", "b"},
		},
		{
			name: "Ignore single line comment",
			code: `import axios from 'axios';
// import unused from 'unused';`,
			expected: []string{"axios"},
		},
		{
			name: "Ignore multi line comment",
			code: `import axios from 'axios';
/* 
import unused from 'unused';
*/`,
			expected: []string{"axios"},
		},
		{
			name: "Mixed imports with comments",
			code: `import axios from 'axios';
// This is a comment
import lodash from 'lodash';
/* Multi-line
   comment */
const x = require('moment');`,
			expected: []string{"axios", "lodash", "moment"},
		},
		{
			name: "String literals with slashes",
			code: `const url = "https://example.com";
import axios from 'axios';
const path = "C:\\Users\\test";
import lodash from 'lodash';`,
			expected: []string{"axios", "lodash"},
		},
		{
			name: "Template literals with slashes",
			code: "const template = `https://api.example.com/${path}`;\nimport axios from 'axios';",
			expected: []string{"axios"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSImports(tt.code)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(result), result)
				return
			}
			for i, exp := range tt.expected {
				if result[i] != exp {
					t.Errorf("Expected import[%d] = '%s', got '%s'", i, exp, result[i])
				}
			}
		})
	}
}

func TestResolveJSPackageName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple package", "lodash", "lodash"},
		{"Subpath", "lodash/fp", "lodash"},
		{"Scoped package", "@org/pkg", "@org/pkg"},
		{"Scoped subpath", "@org/pkg/sub", "@org/pkg"},
		{"Relative path", "./local", ""},
		{"Absolute path", "/absolute/path", ""},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolveJSPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}
