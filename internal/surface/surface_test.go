package surface

import (
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
		if result.Criticality != "Low" {
			t.Errorf("Expected Low criticality for axios, got %s", result.Criticality)
		}
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
	results, err := AnalyzeSurface(tmpDir, deps, nil)
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
	tests := []struct {
		name      string
		fileCount int
		refCount  int
		expected  string
	}{
		{"Low - 1 file, 1 ref", 1, 1, "Low"},
		{"Low - 2 files, 5 refs", 2, 5, "Low"},
		{"Medium - 3 files, 10 refs", 3, 10, "Medium"},
		{"Medium - 9 files, 49 refs", 9, 49, "Medium"},
		{"High - 10 files, 50 refs", 10, 50, "High"},
		{"High - 20 files, 100 refs", 20, 100, "High"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SurfaceResult{
				Files:    make([]string, tt.fileCount),
				RefCount: tt.refCount,
			}
			criticality := calculateCriticality(result)
			if criticality != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, criticality)
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
