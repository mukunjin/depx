package usage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mukunjin/depx/internal/manifest"
)

func TestJSAnalyzer(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, dep := range deps {
		if !results[dep].Used {
			t.Errorf("Expected dependency '%s' to be used", dep)
		}
	}

	results["unused"] = &manifest.UsageResult{Package: "unused", Used: false}
	if results["unused"].Used {
		t.Error("Expected 'unused' to not be used")
	}
}

func TestJSAnalyzerSubpath(t *testing.T) {
	t.Parallel()
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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["lodash"].Used {
		t.Error("Expected 'lodash' to be used via subpath")
	}
	if !results["axios"].Used {
		t.Error("Expected 'axios' to be used via subpath")
	}
}

func TestJSAnalyzerSkipDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

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

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results["axios"].Used {
		t.Error("Expected 'axios' to not be used (should skip node_modules)")
	}
}

func TestJSAnalyzerSkipAllDirs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	skipDirs := []string{"node_modules", "dist", "build", ".git", ".next", "coverage", ".nuxt"}
	for _, dir := range skipDirs {
		subDir := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		jsCode := `import axios from 'axios';`
		if err := os.WriteFile(filepath.Join(subDir, "test.js"), []byte(jsCode), 0644); err != nil {
			t.Fatal(err)
		}
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios"}

	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results["axios"].Used {
		t.Error("All skip dirs should be ignored")
	}
}

func TestJSAnalyzerFileExtensions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		ext  string
	}{
		{"JS", ".js"},
		{"TS", ".ts"},
		{"JSX", ".jsx"},
		{"TSX", ".tsx"},
		{"MJS", ".mjs"},
		{"CJS", ".cjs"},
		{"Vue", ".vue"},
		{"Svelte", ".svelte"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			jsCode := `import axios from 'axios';`
			if err := os.WriteFile(filepath.Join(tmpDir, "test"+tt.ext), []byte(jsCode), 0644); err != nil {
				t.Fatal(err)
			}

			analyzer := NewJSAnalyzer()
			results, err := analyzer.Analyze(tmpDir, []string{"axios"}, nil)
			if err != nil {
				t.Fatalf("Analyze failed: %v", err)
			}

			if !results["axios"].Used {
				t.Errorf("Should detect axios in %s file", tt.ext)
			}
		})
	}
}

func TestJSAnalyzerIgnoredExtensions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	ignoredFiles := []string{"style.css", "data.json", "readme.md", "image.png", "config.yaml"}
	for _, f := range ignoredFiles {
		content := `import axios from 'axios';`
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"}, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results["axios"].Used {
		t.Error("Should not detect imports in non-source files")
	}
}

func TestJSAnalyzerNonExistentDir(t *testing.T) {
	t.Parallel()
	analyzer := NewJSAnalyzer()
	_, err := analyzer.Analyze("/non/existent/dir", []string{"axios"}, nil)
	if err == nil {
		t.Error("Should return error for non-existent directory")
	}
}

func TestJSAnalyzerEmptyDeps(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	jsCode := `import axios from 'axios';`
	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{}, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestJSAnalyzerCommentInCode(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	jsCode := `
import axios from 'axios';
// import unused from 'unused-single';
/* import unused2 from 'unused-multi'; */
import lodash from 'lodash';
/*
import unused3 from 'unused-block';
import unused4 from 'unused-block2';
*/
import react from 'react';
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "react", "unused-single", "unused-multi", "unused-block", "unused-block2"}
	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, dep := range []string{"axios", "lodash", "react"} {
		if !results[dep].Used {
			t.Errorf("Expected '%s' to be used", dep)
		}
	}
	for _, dep := range []string{"unused-single", "unused-multi", "unused-block", "unused-block2"} {
		if results[dep].Used {
			t.Errorf("Expected '%s' to NOT be used (it's in a comment)", dep)
		}
	}
}

func TestJSAnalyzerStringProtection(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	jsCode := `
const url = "import axios from 'fake-axios'";
import realAxios from 'axios';
const path = 'import lodash from "fake-lodash"';
import realLodash from 'lodash';
const template = ` + "`import react from 'fake-react'`" + `;
import realReact from 'react';
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	deps := []string{"axios", "lodash", "react", "fake-axios", "fake-lodash", "fake-react"}
	results, err := analyzer.Analyze(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	for _, dep := range []string{"axios", "lodash", "react"} {
		if !results[dep].Used {
			t.Errorf("Expected '%s' to be used", dep)
		}
	}
	for _, dep := range []string{"fake-axios", "fake-lodash", "fake-react"} {
		if results[dep].Used {
			t.Errorf("Expected '%s' to NOT be used (it's inside a string)", dep)
		}
	}
}

func TestJSAnalyzerRefCount(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	jsCode := `
import axios from 'axios';
import { get } from 'axios';
const axios2 = require('axios');
`

	if err := os.WriteFile(filepath.Join(tmpDir, "test.js"), []byte(jsCode), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios"}, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if results["axios"].RefCount < 3 {
		t.Errorf("Expected RefCount >= 3, got %d", results["axios"].RefCount)
	}
}

func TestJSAnalyzerMultiFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "src", "components")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(`import axios from 'axios';`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "App.tsx"), []byte(`import React from 'react';\nimport axios from 'axios';`), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewJSAnalyzer()
	results, err := analyzer.Analyze(tmpDir, []string{"axios", "react"}, nil)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if !results["axios"].Used {
		t.Error("axios should be used")
	}
	if !results["react"].Used {
		t.Error("react should be used")
	}
	// axios is used in 2 files
	if len(results["axios"].UsedIn) != 2 {
		t.Errorf("axios should be used in 2 files, got %d", len(results["axios"].UsedIn))
	}
	// react is used in 1 file
	if len(results["react"].UsedIn) != 1 {
		t.Errorf("react should be used in 1 file, got %d", len(results["react"].UsedIn))
	}
}

func TestExtractJSImports(t *testing.T) {
	t.Parallel()
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
			name:     "Template literals with slashes",
			code:     "const template = `https://api.example.com/${path}`;\nimport axios from 'axios';",
			expected: []string{"axios"},
		},
		{
			name:     "Import with double quotes",
			code:     `import axios from "axios";`,
			expected: []string{"axios"},
		},
		{
			name:     "Require with double quotes",
			code:     `const fs = require("fs");`,
			expected: []string{"fs"},
		},
		{
			name:     "Namespace import",
			code:     `import * as path from 'path';`,
			expected: []string{"path"},
		},
		{
			name:     "Default and named import",
			code:     `import React, { useState, useEffect } from 'react';`,
			expected: []string{"react"},
		},
		{
			name:     "Export all from",
			code:     `export * from 'module';`,
			expected: []string{"module"},
		},
		{
			name: "Import in string should be ignored",
			code: `const s = "import fake from 'fake-pkg'";
import real from 'real-pkg';`,
			expected: []string{"real-pkg"},
		},
		{
			name:     "Comment at end of import line",
			code:     `import axios from 'axios'; // http client`,
			expected: []string{"axios"},
		},
		{
			name: "Multiple comments in block",
			code: `/* comment 1 */
import a from 'a';
/* comment 2 */
import b from 'b';
/* comment 3 */`,
			expected: []string{"a", "b"},
		},
		{
			name:     "Empty string",
			code:     "",
			expected: nil,
		},
		{
			name:     "Only comments",
			code:     "// just a comment\n/* another comment */",
			expected: nil,
		},
		{
			name: "Escaped quote in string",
			code: `const s = "import fake from 'fake'";
import real from 'real';`,
			expected: []string{"real"},
		},
		{
			name:     "Import with semicolon",
			code:     `import axios from 'axios';`,
			expected: []string{"axios"},
		},
		{
			name:     "Import without semicolon",
			code:     `import axios from 'axios'`,
			expected: []string{"axios"},
		},
		{
			name:     "Destructured require",
			code:     `const { get, post } = require('axios');`,
			expected: []string{"axios"},
		},
		{
			name:     "Nested template literal",
			code:     "const s = `outer ${`inner`} outer`;\nimport axios from 'axios';",
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
		{"Relative parent", "../parent", ""},
		{"Absolute path", "/absolute/path", ""},
		{"Empty string", "", ""},
		{"Deep subpath", "lodash/fp/deep/nested", "lodash"},
		{"Scoped deep subpath", "@org/pkg/a/b/c", "@org/pkg"},
		{"Single scoped", "@org/pkg", "@org/pkg"},
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

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comments",
			input:    `import axios from 'axios';`,
			expected: `import axios from 'axios';`,
		},
		{
			name:     "Single line comment",
			input:    "// comment\nimport axios from 'axios';",
			expected: "\nimport axios from 'axios';",
		},
		{
			name:     "Multi line comment",
			input:    "/* comment */\nimport axios from 'axios';",
			expected: "\nimport axios from 'axios';",
		},
		{
			name:     "Comment inside string double quote",
			input:    `const s = "// not a comment";`,
			expected: `const s = "// not a comment";`,
		},
		{
			name:     "Comment inside string single quote",
			input:    `const s = '// not a comment';`,
			expected: `const s = '// not a comment';`,
		},
		{
			name:     "Comment inside template literal",
			input:    "const s = `// not a comment`;",
			expected: "const s = `// not a comment`;",
		},
		{
			name:     "Block comment inside string",
			input:    `const s = "/* not a comment */";`,
			expected: `const s = "/* not a comment */";`,
		},
		{
			name:     "Escaped quote in string",
			input:    `const s = "escaped \" // still string";`,
			expected: `const s = "escaped \" // still string";`,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only single line comment",
			input:    "// just a comment",
			expected: "",
		},
		{
			name:     "Only multi line comment",
			input:    "/* just a comment */",
			expected: "",
		},
		{
			name:     "Mixed content",
			input:    "import a from 'a';\n// comment\nimport b from 'b';\n/* block */\nimport c from 'c';",
			expected: "import a from 'a';\n\nimport b from 'b';\n\nimport c from 'c';",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeComments(tt.input)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}
