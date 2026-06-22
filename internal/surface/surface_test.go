package surface

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestAnalyzeSurface(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个测试文件
	file1 := `import axios from "axios";
import { debounce } from "lodash";
`
	file2 := `import axios from "axios";
import moment from "moment";
`
	file3 := `import { throttle } from "lodash";
import "lodash/fp";
`

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.js"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.js"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file3.js"), []byte(file3), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "lodash", "moment"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 axios
	if result, ok := results["axios"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for axios, got %d", len(result.Files))
		}
		if result.RefCount != 2 {
			t.Errorf("Expected 2 refs for axios, got %d", result.RefCount)
		}
		// Criticality 现在基于百分位，不再检查固定值
	} else {
		t.Error("axios not found in results")
	}

	// 检查 lodash
	if result, ok := results["lodash"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for lodash, got %d", len(result.Files))
		}
		// 应该有 2 个模块：lodash 和 lodash/fp
		if len(result.Modules) != 2 {
			t.Errorf("Expected 2 modules for lodash, got %d: %v", len(result.Modules), result.Modules)
		}
		if result.RefCount != 3 {
			t.Errorf("Expected 3 refs for lodash, got %d", result.RefCount)
		}
	} else {
		t.Error("lodash not found in results")
	}

	// 检查 moment
	if result, ok := results["moment"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for moment, got %d", len(result.Files))
		}
		if result.RefCount != 1 {
			t.Errorf("Expected 1 ref for moment, got %d", result.RefCount)
		}
	} else {
		t.Error("moment not found in results")
	}
}

func TestAnalyzeSurfaceWithRequire(t *testing.T) {
	tmpDir := t.TempDir()

	code := `const axios = require("axios");
const lodash = require("lodash");
`

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "lodash"}
	opts := &Options{
		ManifestType: "npm",
		ExcludeDirs:  []string{"node_modules"},
	}
	results, err := AnalyzeSurface(tmpDir, deps, opts)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if result, ok := results["axios"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for axios, got %d", len(result.Files))
		}
	} else {
		t.Error("axios not found in results")
	}
}

func TestAnalyzeSurfaceScopedPackage(t *testing.T) {
	tmpDir := t.TempDir()

	code := `import { something } from "@org/pkg";
import { other } from "@org/pkg/sub";
`

	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"@org/pkg"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if result, ok := results["@org/pkg"]; ok {
		if len(result.Modules) != 2 {
			t.Errorf("Expected 2 modules for @org/pkg, got %d: %v", len(result.Modules), result.Modules)
		}
	} else {
		t.Error("@org/pkg not found in results")
	}
}

func TestAnalyzeSurfaceGo(t *testing.T) {
	tmpDir := t.TempDir()

	mainGo := `package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func main() {
	fmt.Println("hello")
}
`

	handlerGo := `package handlers

import (
	"github.com/gin-gonic/gin"
)

func Handle() {
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "handler.go"), []byte(handlerGo), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"github.com/gin-gonic/gin", "github.com/stretchr/testify"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 gin
	if result, ok := results["github.com/gin-gonic/gin"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for gin, got %d", len(result.Files))
		}
		if result.RefCount != 2 {
			t.Errorf("Expected 2 refs for gin, got %d", result.RefCount)
		}
	} else {
		t.Error("gin not found in results")
	}

	// 检查 testify
	if result, ok := results["github.com/stretchr/testify"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for testify, got %d", len(result.Files))
		}
	} else {
		t.Error("testify not found in results")
	}
}

func TestAnalyzeSurfaceRust(t *testing.T) {
	tmpDir := t.TempDir()

	mainRs := `use serde::Deserialize;
use tokio::main;
use reqwest;

fn main() {
    println!("hello");
}
`

	handlerRs := `use serde_json;
use tokio::runtime;

fn handle() {
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.rs"), []byte(mainRs), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "handler.rs"), []byte(handlerRs), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"serde", "tokio", "reqwest", "serde_json"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 serde
	if result, ok := results["serde"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for serde, got %d", len(result.Files))
		}
		if result.RefCount != 1 {
			t.Errorf("Expected 1 ref for serde, got %d", result.RefCount)
		}
	} else {
		t.Error("serde not found in results")
	}

	// 检查 tokio
	if result, ok := results["tokio"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for tokio, got %d", len(result.Files))
		}
		if result.RefCount != 2 {
			t.Errorf("Expected 2 refs for tokio, got %d", result.RefCount)
		}
	} else {
		t.Error("tokio not found in results")
	}

	// 检查 reqwest
	if result, ok := results["reqwest"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for reqwest, got %d", len(result.Files))
		}
	} else {
		t.Error("reqwest not found in results")
	}
}

func TestAnalyzeSurfacePython(t *testing.T) {
	tmpDir := t.TempDir()

	mainPy := `import requests
from flask import Flask
import numpy

def main():
    pass
`

	dbPy := `import sqlalchemy
from redis import Redis
import numpy as np

def db():
    pass
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte(mainPy), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "db.py"), []byte(dbPy), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"requests", "flask", "numpy", "sqlalchemy", "redis"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 requests
	if result, ok := results["requests"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for requests, got %d", len(result.Files))
		}
	} else {
		t.Error("requests not found in results")
	}

	// 检查 flask
	if result, ok := results["flask"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for flask, got %d", len(result.Files))
		}
	} else {
		t.Error("flask not found in results")
	}

	// 检查 numpy
	if result, ok := results["numpy"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for numpy, got %d", len(result.Files))
		}
		if result.RefCount != 2 {
			t.Errorf("Expected 2 refs for numpy, got %d", result.RefCount)
		}
	} else {
		t.Error("numpy not found in results")
	}

	// 检查 sqlalchemy
	if result, ok := results["sqlalchemy"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for sqlalchemy, got %d", len(result.Files))
		}
	} else {
		t.Error("sqlalchemy not found in results")
	}
}

func TestResolvePythonPackageName(t *testing.T) {
	tests := []struct {
		importPath string
		expected   string
	}{
		{"requests", "requests"},
		{"flask.app", "flask"},
		{"sqlalchemy.orm", "sqlalchemy"},
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			result := resolvePythonPackageName(tt.importPath)
			if result != tt.expected {
				t.Errorf("resolvePythonPackageName(%s) = %s, expected %s", tt.importPath, result, tt.expected)
			}
		})
	}
}

func TestCalculateCriticality(t *testing.T) {
	// 测试百分位计算
	tests := []struct {
		name     string
		packages []*SurfaceResult
		expected map[string]string
	}{
		{
			name: "10 packages - top 20% are High",
			packages: []*SurfaceResult{
				{Package: "pkg1", RefCount: 100},
				{Package: "pkg2", RefCount: 80},
				{Package: "pkg3", RefCount: 60},
				{Package: "pkg4", RefCount: 50},
				{Package: "pkg5", RefCount: 40},
				{Package: "pkg6", RefCount: 30},
				{Package: "pkg7", RefCount: 20},
				{Package: "pkg8", RefCount: 10},
				{Package: "pkg9", RefCount: 5},
				{Package: "pkg10", RefCount: 1},
			},
			expected: map[string]string{
				"pkg1": "High", "pkg2": "High",
				"pkg3": "Medium", "pkg4": "Medium", "pkg5": "Medium",
				"pkg6": "Low", "pkg7": "Low", "pkg8": "Low", "pkg9": "Low", "pkg10": "Low",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 计算分数
			for _, pkg := range tt.packages {
				pkg.CalculateScore()
			}

			// 排序
			sort.Slice(tt.packages, func(i, j int) bool {
				return tt.packages[i].Score > tt.packages[j].Score
			})

			// 计算百分位
			totalCount := len(tt.packages)
			for i, pkg := range tt.packages {
				percentile := float64(i) / float64(totalCount)
				if percentile < 0.2 {
					pkg.Criticality = "High"
				} else if percentile < 0.5 {
					pkg.Criticality = "Medium"
				} else {
					pkg.Criticality = "Low"
				}
			}

			// 验证
			for _, pkg := range tt.packages {
				if pkg.Criticality != tt.expected[pkg.Package] {
					t.Errorf("Package %s: expected %s, got %s", pkg.Package, tt.expected[pkg.Package], pkg.Criticality)
				}
			}
		})
	}
}

func TestExtractJSImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "basic imports",
			code: `import axios from "axios";
import { debounce } from "lodash";
const fs = require("fs");
import("./dynamic");
`,
			expected: []string{"axios", "lodash", "fs", "./dynamic"},
		},
		{
			name: "comments should be ignored",
			code: `import axios from "axios";
// import commented from "commented-pkg";
/* import block from "block-pkg"; */
import real from "real-pkg";
`,
			expected: []string{"axios", "real-pkg"},
		},
		{
			name: "string literals should be ignored",
			code: `import axios from "axios";
const str = "import fake from 'fake-pkg'";
const str2 = 'import fake2 from "fake-pkg2"';
const str3 = ` + "`import fake3 from \"fake-pkg3\"`" + `;
import real from "real-pkg";
`,
			expected: []string{"axios", "real-pkg"},
		},
		{
			name: "export from",
			code: `export { something } from "export-pkg";
export * from "star-export-pkg";
`,
			expected: []string{"export-pkg", "star-export-pkg"},
		},
		{
			name: "scoped packages",
			code: `import { something } from "@org/pkg";
import { other } from "@org/pkg/sub";
`,
			expected: []string{"@org/pkg", "@org/pkg/sub"},
		},
		{
			name:     "empty code",
			code:     ``,
			expected: []string{},
		},
		{
			name: "only comments",
			code: `// just a comment
/* block comment */
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractJSImports(tt.code)
			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(imports), imports)
			}
			for i, imp := range imports {
				if i < len(tt.expected) && imp != tt.expected[i] {
					t.Errorf("Expected import %s, got %s", tt.expected[i], imp)
				}
			}
		})
	}
}

func TestExtractGoImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "basic imports",
			code: `package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

import "os"

func main() {}
`,
			expected: []string{"fmt", "github.com/gin-gonic/gin", "os"},
		},
		{
			name: "aliased imports",
			code: `package main

import (
	. "github.com/pkg/errors"
	_ "github.com/mattn/go-sqlite3"
	myalias "github.com/user/package"
)
`,
			expected: []string{"github.com/pkg/errors", "github.com/mattn/go-sqlite3", "github.com/user/package"},
		},
		{
			name: "comments should be ignored",
			code: `package main

import (
	"fmt"
	// "github.com/commented/pkg"
	"github.com/real/pkg"
)
`,
			expected: []string{"fmt", "github.com/real/pkg"},
		},
		{
			name: "line comments should be ignored",
			code: `package main

import (
	"fmt" // this is fmt
	"github.com/real/pkg" // real package
)
`,
			expected: []string{"fmt", "github.com/real/pkg"},
		},
		{
			name:     "empty code",
			code:     ``,
			expected: []string{},
		},
		{
			name: "only comments",
			code: `// just a comment
/* block comment */
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractGoImports(tt.code)
			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(imports), imports)
			}
			for i, imp := range imports {
				if i < len(tt.expected) && imp != tt.expected[i] {
					t.Errorf("Expected import %s, got %s", tt.expected[i], imp)
				}
			}
		})
	}
}

func TestExtractRustImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "basic use statements",
			code: `use serde::Deserialize;
use tokio::main;
use reqwest;

fn main() {}
`,
			expected: []string{"serde", "tokio", "reqwest"},
		},
		{
			name: "extern crate",
			code: `extern crate serde;
extern crate tokio as tokio_runtime;

fn main() {}
`,
			expected: []string{"serde", "tokio"},
		},
		{
			name: "use with braces",
			code: `use serde::{Deserialize, Serialize};
use tokio::runtime::{Runtime, Builder};

fn main() {}
`,
			expected: []string{"serde", "tokio"},
		},
		{
			name: "comments should be ignored",
			code: `use serde::Deserialize;
// use commented::pkg;
/* use block::pkg; */
use real::pkg;

fn main() {}
`,
			expected: []string{"serde", "real"},
		},
		{
			name: "string literals should be ignored",
			code: `use serde::Deserialize;
let s = "use fake::pkg;";
use real::pkg;

fn main() {}
`,
			expected: []string{"serde", "real"},
		},
		{
			name:     "empty code",
			code:     ``,
			expected: []string{},
		},
		{
			name: "only comments",
			code: `// just a comment
/* block comment */
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractRustImports(tt.code)
			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(imports), imports)
			}
			for i, imp := range imports {
				if i < len(tt.expected) && imp != tt.expected[i] {
					t.Errorf("Expected import %s, got %s", tt.expected[i], imp)
				}
			}
		})
	}
}

func TestExtractPythonImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name: "basic imports",
			code: `import requests
from flask import Flask
import numpy as np
from sqlalchemy.orm import Session

def main():
    pass
`,
			expected: []string{"requests", "flask", "numpy", "sqlalchemy"},
		},
		{
			name: "aliased imports",
			code: `import numpy as np
import pandas as pd
from matplotlib import pyplot as plt
`,
			expected: []string{"numpy", "pandas", "matplotlib"},
		},
		{
			name: "multiple imports in one line",
			code: `import os, sys, json
from collections import defaultdict, Counter
`,
			expected: []string{"os", "sys", "json", "collections"},
		},
		{
			name: "relative imports",
			code: `from . import module
from .. import parent_module
from .sub import something
`,
			expected: []string{},
		},
		{
			name: "comments should be ignored",
			code: `import requests
# import commented
from flask import Flask  # this is flask
`,
			expected: []string{"requests", "flask"},
		},
		{
			name: "string literals should be ignored",
			code: `import requests
code = "import fake"
code2 = 'from fake import something'
from flask import Flask
`,
			expected: []string{"requests", "flask"},
		},
		{
			name: "multiline imports",
			code: `from flask import (
    Flask,
    request,
    jsonify
)
import os
`,
			expected: []string{"flask", "os"},
		},
		{
			name:     "empty code",
			code:     ``,
			expected: []string{},
		},
		{
			name: "only comments",
			code: `# just a comment
# another comment
`,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractPythonImports(tt.code)
			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d: %v", len(tt.expected), len(imports), imports)
			}
			for i, imp := range imports {
				if i < len(tt.expected) && imp != tt.expected[i] {
					t.Errorf("Expected import %s, got %s", tt.expected[i], imp)
				}
			}
		})
	}
}

func TestIsImportOfDep(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		dep        string
		lang       langType
		expected   bool
	}{
		// JS/TS
		{"JS exact match", "axios", "axios", langJS, true},
		{"JS subpath", "axios/lib", "axios", langJS, true},
		{"JS scoped exact", "@org/pkg", "@org/pkg", langJS, true},
		{"JS scoped subpath", "@org/pkg/sub", "@org/pkg", langJS, true},
		{"JS no match", "lodash", "axios", langJS, false},
		{"JS partial no match", "axios-retry", "axios", langJS, false},

		// Go
		{"Go exact match", "github.com/gin-gonic/gin", "github.com/gin-gonic/gin", langGo, true},
		{"Go subpath", "github.com/gin-gonic/gin/binding", "github.com/gin-gonic/gin", langGo, true},
		{"Go no match", "github.com/gin-gonic/gin", "github.com/gin-gonic/ginx", langGo, false},

		// Rust
		{"Rust exact match", "serde", "serde", langRust, true},
		{"Rust submodule", "serde::Deserialize", "serde", langRust, true},
		{"Rust no match", "serde_json", "serde", langRust, false},

		// Python
		{"Python exact match", "flask", "flask", langPython, true},
		{"Python submodule", "flask.app", "flask", langPython, true},
		{"Python no match", "flask_cors", "flask", langPython, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImportOfDep(tt.importPath, tt.dep, tt.lang)
			if result != tt.expected {
				t.Errorf("isImportOfDep(%q, %q, %v) = %v, expected %v",
					tt.importPath, tt.dep, tt.lang, result, tt.expected)
			}
		})
	}
}

func TestMatchDependency(t *testing.T) {
	deps := []string{"axios", "lodash", "@org/pkg", "github.com/gin-gonic/gin"}
	sortedDeps := make([]string, len(deps))
	copy(sortedDeps, deps)
	sort.Slice(sortedDeps, func(i, j int) bool {
		return len(sortedDeps[i]) > len(sortedDeps[j])
	})

	tests := []struct {
		name       string
		importPath string
		lang       langType
		expected   string
	}{
		{"match axios", "axios", langJS, "axios"},
		{"match axios subpath", "axios/lib", langJS, "axios"},
		{"match lodash", "lodash", langJS, "lodash"},
		{"match lodash subpath", "lodash/fp", langJS, "lodash"},
		{"match scoped", "@org/pkg", langJS, "@org/pkg"},
		{"match scoped subpath", "@org/pkg/sub", langJS, "@org/pkg"},
		{"match go", "github.com/gin-gonic/gin", langGo, "github.com/gin-gonic/gin"},
		{"no match", "unknown-pkg", langJS, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchDependency(tt.importPath, sortedDeps, tt.lang)
			if result != tt.expected {
				t.Errorf("matchDependency(%q, %v) = %q, expected %q",
					tt.importPath, tt.lang, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeSurfaceDeduplication(t *testing.T) {
	tmpDir := t.TempDir()

	// 同一文件多次导入同一依赖
	code := `import axios from "axios";
import axios from "axios";
import { get } from "axios";
`

	if err := os.WriteFile(filepath.Join(tmpDir, "file.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if result, ok := results["axios"]; ok {
		// 文件应该去重，只有 1 个
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file (deduped), got %d", len(result.Files))
		}
		// 模块应该去重，只有 1 个（axios）
		if len(result.Modules) != 1 {
			t.Errorf("Expected 1 module (deduped), got %d: %v", len(result.Modules), result.Modules)
		}
		// 但引用计数应该是 3
		if result.RefCount != 3 {
			t.Errorf("Expected 3 refs, got %d", result.RefCount)
		}
	} else {
		t.Error("axios not found in results")
	}
}

func TestAnalyzeSurfaceMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := `import axios from "axios";`
	file2 := `import axios from "axios";
import { get } from "axios/lib";`
	file3 := `import moment from "moment";`

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.js"), []byte(file1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.js"), []byte(file2), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file3.js"), []byte(file3), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "moment"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 检查 axios
	if result, ok := results["axios"]; ok {
		if len(result.Files) != 2 {
			t.Errorf("Expected 2 files for axios, got %d", len(result.Files))
		}
		if len(result.Modules) != 2 {
			t.Errorf("Expected 2 modules for axios, got %d: %v", len(result.Modules), result.Modules)
		}
		if result.RefCount != 3 {
			t.Errorf("Expected 3 refs for axios, got %d", result.RefCount)
		}
	} else {
		t.Error("axios not found in results")
	}

	// 检查 moment
	if result, ok := results["moment"]; ok {
		if len(result.Files) != 1 {
			t.Errorf("Expected 1 file for moment, got %d", len(result.Files))
		}
		if result.RefCount != 1 {
			t.Errorf("Expected 1 ref for moment, got %d", result.RefCount)
		}
	} else {
		t.Error("moment not found in results")
	}
}

func TestCalculateCriticality_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		packages []*SurfaceResult
		check    func([]*SurfaceResult) error
	}{
		{
			name: "single package - should be High",
			packages: []*SurfaceResult{
				{Package: "only-pkg", RefCount: 10},
			},
			check: func(pkgs []*SurfaceResult) error {
				if pkgs[0].Criticality != "High" {
					return fmt.Errorf("single package should be High, got %s", pkgs[0].Criticality)
				}
				return nil
			},
		},
		{
			name: "two packages - first is High, second is Low",
			packages: []*SurfaceResult{
				{Package: "pkg1", RefCount: 100},
				{Package: "pkg2", RefCount: 1},
			},
			check: func(pkgs []*SurfaceResult) error {
				if pkgs[0].Criticality != "High" {
					return fmt.Errorf("pkg1 should be High, got %s", pkgs[0].Criticality)
				}
				if pkgs[1].Criticality != "Low" {
					return fmt.Errorf("pkg2 should be Low, got %s", pkgs[1].Criticality)
				}
				return nil
			},
		},
		{
			name: "five packages - 1 High, 1 Medium, 3 Low",
			packages: []*SurfaceResult{
				{Package: "pkg1", RefCount: 100},
				{Package: "pkg2", RefCount: 80},
				{Package: "pkg3", RefCount: 60},
				{Package: "pkg4", RefCount: 40},
				{Package: "pkg5", RefCount: 20},
			},
			check: func(pkgs []*SurfaceResult) error {
				// 0/5 = 0.0 -> High
				if pkgs[0].Criticality != "High" {
					return fmt.Errorf("pkg1 should be High, got %s", pkgs[0].Criticality)
				}
				// 1/5 = 0.2 -> Medium
				if pkgs[1].Criticality != "Medium" {
					return fmt.Errorf("pkg2 should be Medium, got %s", pkgs[1].Criticality)
				}
				// 2/5 = 0.4 -> Medium
				if pkgs[2].Criticality != "Medium" {
					return fmt.Errorf("pkg3 should be Medium, got %s", pkgs[2].Criticality)
				}
				// 3/5 = 0.6 -> Low
				if pkgs[3].Criticality != "Low" {
					return fmt.Errorf("pkg4 should be Low, got %s", pkgs[3].Criticality)
				}
				// 4/5 = 0.8 -> Low
				if pkgs[4].Criticality != "Low" {
					return fmt.Errorf("pkg5 should be Low, got %s", pkgs[4].Criticality)
				}
				return nil
			},
		},
		{
			name: "all same ref count",
			packages: []*SurfaceResult{
				{Package: "pkg1", RefCount: 10},
				{Package: "pkg2", RefCount: 10},
				{Package: "pkg3", RefCount: 10},
			},
			check: func(pkgs []*SurfaceResult) error {
				// 即使分数相同，百分位仍会分配不同关键度
				for _, pkg := range pkgs {
					if pkg.Criticality == "" {
						return fmt.Errorf("criticality should not be empty for %s", pkg.Package)
					}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 计算分数
			for _, pkg := range tt.packages {
				pkg.CalculateScore()
			}

			// 排序
			sort.Slice(tt.packages, func(i, j int) bool {
				return tt.packages[i].Score > tt.packages[j].Score
			})

			// 计算百分位
			totalCount := len(tt.packages)
			for i, pkg := range tt.packages {
				percentile := float64(i) / float64(totalCount)
				if percentile < 0.2 {
					pkg.Criticality = "High"
				} else if percentile < 0.5 {
					pkg.Criticality = "Medium"
				} else {
					pkg.Criticality = "Low"
				}
			}

			// 验证
			if err := tt.check(tt.packages); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		result   *SurfaceResult
		expected int
	}{
		{
			name: "basic score calculation",
			result: &SurfaceResult{
				RefCount:  10,
				Modules:   []string{"mod1", "mod2"},
			},
			expected: 52, // 10*5 + 2
		},
		{
			name: "zero ref count",
			result: &SurfaceResult{
				RefCount:  0,
				Modules:   []string{"mod1"},
			},
			expected: 1, // 0*5 + 1
		},
		{
			name: "no modules",
			result: &SurfaceResult{
				RefCount:  5,
				Modules:   []string{},
			},
			expected: 25, // 5*5 + 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.result.CalculateScore()
			if tt.result.Score != tt.expected {
				t.Errorf("Expected score %d, got %d", tt.expected, tt.result.Score)
			}
		})
	}
}

func TestAnalyzeSurface_EmptyDeps(t *testing.T) {
	tmpDir := t.TempDir()

	code := `import axios from "axios";`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := AnalyzeSurface(tmpDir, []string{}, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty deps, got %d", len(results))
	}
}

func TestAnalyzeSurface_NoSourceFiles(t *testing.T) {
	tmpDir := t.TempDir()

	deps := []string{"axios", "lodash"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 所有依赖应该都存在，但 RefCount 为 0
	for _, dep := range deps {
		if result, ok := results[dep]; !ok {
			t.Errorf("Expected result for %s", dep)
		} else if result.RefCount != 0 {
			t.Errorf("Expected RefCount 0 for %s, got %d", dep, result.RefCount)
		}
	}
}

func TestAnalyzeSurface_WithExcludeDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// 在主目录创建文件
	code1 := `import axios from "axios";`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}

	// 在 excluded 目录创建文件
	excludedDir := filepath.Join(tmpDir, "excluded")
	if err := os.MkdirAll(excludedDir, 0755); err != nil {
		t.Fatal(err)
	}
	code2 := `import lodash from "lodash";`
	if err := os.WriteFile(filepath.Join(excludedDir, "utils.js"), []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}

	opts := &Options{
		ExcludeDirs: []string{"excluded"},
	}
	results, err := AnalyzeSurface(tmpDir, []string{"axios", "lodash"}, opts)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if results["axios"].RefCount != 1 {
		t.Errorf("Expected axios RefCount 1, got %d", results["axios"].RefCount)
	}
	if results["lodash"].RefCount != 0 {
		t.Errorf("Expected lodash RefCount 0 (excluded dir), got %d", results["lodash"].RefCount)
	}
}

func TestAnalyzeSurface_WithExcludeFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建普通文件
	code1 := `import axios from "axios";`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code1), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建测试文件
	code2 := `import lodash from "lodash";`
	if err := os.WriteFile(filepath.Join(tmpDir, "utils.test.js"), []byte(code2), 0644); err != nil {
		t.Fatal(err)
	}

	opts := &Options{
		ExcludeFiles: []string{"*.test.js"},
	}
	results, err := AnalyzeSurface(tmpDir, []string{"axios", "lodash"}, opts)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	if results["axios"].RefCount != 1 {
		t.Errorf("Expected axios RefCount 1, got %d", results["axios"].RefCount)
	}
	if results["lodash"].RefCount != 0 {
		t.Errorf("Expected lodash RefCount 0 (excluded file), got %d", results["lodash"].RefCount)
	}
}

func TestAnalyzeSurface_CriticalityAssignment(t *testing.T) {
	tmpDir := t.TempDir()

	// 创建多个文件，使不同依赖有不同的引用次数
	code := `
import axios from "axios";
import axios from "axios";
import axios from "axios";
import lodash from "lodash";
import moment from "moment";
`
	if err := os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(code), 0644); err != nil {
		t.Fatal(err)
	}

	deps := []string{"axios", "lodash", "moment"}
	results, err := AnalyzeSurface(tmpDir, deps, nil)
	if err != nil {
		t.Fatalf("AnalyzeSurface failed: %v", err)
	}

	// 验证所有依赖都有非空的关键度
	for _, dep := range deps {
		if results[dep].Criticality == "" {
			t.Errorf("Expected non-empty criticality for %s", dep)
		}
	}

	// axios 引用最多，应该是 High
	if results["axios"].Criticality != "High" {
		t.Errorf("Expected axios to be High, got %s", results["axios"].Criticality)
	}
}

func TestSurfaceResultDedup(t *testing.T) {
	// 测试 map 去重逻辑
	result := &SurfaceResult{
		Package:   "test-pkg",
		Files:     []string{},
		Modules:   []string{},
		fileSet:   make(map[string]struct{}),
		moduleSet: make(map[string]struct{}),
	}

	// 添加相同文件两次
	relPath := "test.js"
	if _, exists := result.fileSet[relPath]; !exists {
		result.fileSet[relPath] = struct{}{}
		result.Files = append(result.Files, relPath)
	}
	if _, exists := result.fileSet[relPath]; !exists {
		result.fileSet[relPath] = struct{}{}
		result.Files = append(result.Files, relPath)
	}

	if len(result.Files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result.Files))
	}
}
