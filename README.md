# depx

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-GPLv3-blue?style=flat-square)](./LICENSE)

Dependency analyzer for modern projects — detect unused dependencies and assess runtime dependency surface area.

**中文文档**: [README_zh.md](README_zh.md)

---

## Features

- **Unused Detection** — Scan manifest files and match declared dependencies against source imports
- **Surface Area Analysis** — Measure how widely runtime dependencies are used (files, modules, ref count, criticality)
- **Lock File Analysis** — Parse lock files for transitive dependency visibility
- **Configuration** — Customize ignore rules and exclusions via `.depx.yml`

## Supported

| Package Manager | Manifest | Lock File | Source Files |
|----------------|----------|-----------|--------------|
| npm | package.json | package-lock.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | go.sum | .go |
| Rust | Cargo.toml | Cargo.lock | .rs |
| Python | requirements.txt | — | .py |

**Not supported yet:** `yarn.lock`, `pnpm-lock.yaml`, `pyproject.toml`, `Pipfile`, npm/pnpm/yarn workspaces member manifests (root manifest only).

## Quick Start

```bash
git clone https://github.com/mukunjin/depx.git
cd depx
go build .
depx scan
depx surface
```

## Installation (Windows)

### First-time Setup

Windows blocks PowerShell scripts by default. Run PowerShell as **Administrator**:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Install via scripts

```powershell
git clone https://github.com/mukunjin/depx.git
cd depx
.\build.ps1
.\install.ps1
```

### Uninstall

```powershell
.\install.ps1 -Uninstall
```

The install script will:
- Copy `depx.exe` to `%LOCALAPPDATA%\depx`
- Add to user PATH
- Verify with `depx --version`

### Version Management

| Location | Role | When it takes effect |
|----------|------|---------------------|
| Git tag | Primary source | When building via `.\build.ps1` |
| `cmd/root.go` line 8 | Fallback (`dev`) | When running `go build` without `-ldflags` |
| `build.ps1` | Reads git tag, injects via `-ldflags` | Every `.\build.ps1` run |
| `install.ps1` line 140 | Reads version from binary | During installation verification |

**Release a new version:**
1. `git tag v0.3.0`
2. `git push origin v0.3.0`
3. `.\build.ps1`
4. `.\install.ps1`

## Usage

### `scan` vs `surface`

| | `depx scan` | `depx surface` |
|---|-------------|----------------|
| **Purpose** | Find unused dependencies | Measure usage breadth of dependencies |
| **Scope** | `dependencies` + `devDependencies` | `dependencies` only (use `--dev` for dev) |
| **Output** | Runtime / Tool / Unused lists | Files, modules, ref count, criticality |
| **Indirect deps** | Summary with `--indirect`, full with `--indirect-all` | Shared transitive summary with `--indirect` |
| **Type packages** | Separate `@types/*` stats | Excluded entirely |

### Scan

```bash
depx scan
depx scan /path/to/project
depx scan --config /path/to/.depx.yml
depx scan --indirect    # show indirect deps summary (Total + Top Shared)
depx scan -i
depx scan --indirect-all  # show all indirect dependencies
depx scan --types       # show @types/* packages
depx scan -t
```

`scan` checks **both** `dependencies` and `devDependencies` against source imports.

Example output:

```
  Project Summary
--------------------------
  Path:            /path/to/project
  Package Manager: npm
  Dependencies:    12
  Used:            7
  Unused:          5
  Type Packages:   3 (use --types to show)
  Indirect:        358 (use --indirect to show)

  Runtime Dependencies
--------------------------
  [✓] express
  [✓] lodash
  [✓] axios

  Tool Packages
--------------------------
  [✓] jest
  [✓] eslint

  Unused Dependencies
--------------------------
  [x] moment
  [x] chalk
```

With `--indirect` flag:

```
  Indirect Dependencies
--------------------------
  Total: 358

  Top Shared
  --------------------------
  clsx                 (4 parents)
  scheduler            (3 parents)
  redux                (2 parents)

  Use --indirect-all to show all packages.
```

**Notes:**
- `@types/*` packages are counted separately and excluded from Runtime/Tool/Unused lists
- **Runtime Dependencies**: packages from `dependencies` (production)
- **Tool Packages**: packages from `devDependencies` (build tools, test frameworks, etc.)
- Enable `lock_file: true` in `.depx.yml` (default) to see indirect dependency counts

### Surface Area Analysis

```bash
depx surface              # runtime dependencies only
depx surface --dev        # include devDependencies
depx surface -D
depx surface --indirect   # shared transitive dependency summary
depx surface -i
```

Example output:

```
  Surface Area
===================================

  Summary
--------------------------
  Packages:     10
  High:         2
  Medium:       3
  Low:          5

  Most Critical
--------------------------
  @mui/material
  Score: 64

  Runtime Surface
--------------------------
  @mui/material
    Criticality: High
    Files: 43
    Modules: 8
    Ref Count: 182

  chalk
    Criticality: Low
    Files: 4
    Modules: 1
    Ref Count: 6

  Indirect Packages
--------------------------

  Total: 531

  Top Shared Dependencies
--------------------------

  clsx
    Required By: 4 direct packages

  scheduler
    Required By: 3 direct packages
```

**Notes:**
- Default scope is **runtime surface** (`dependencies` only)
- **Score** = `RefCount × 5 + Modules` — simple, no double-counting
- **Criticality** uses **percentile ranking** — Top 20% = High, Top 50% = Medium, rest = Low. Any project always has a clear hierarchy regardless of size
- `--indirect` shows total transitive count + packages required by **2+ direct packages**
- Shared indirect graph requires `package-lock.json` (npm lockfile v2+ recommended)
- `@types/*` are excluded

## Configuration

Create `.depx.yml` in your project root:

```yaml
ignore:
  - "@types/node"
  - "typescript"

exclude_dirs:
  - "vendor"
  - "dist"
  - "node_modules"

exclude_files:
  - "*.test.js"
  - "*.spec.ts"

read_node_modules: false
lock_file: true
```

| Key | Default | Description |
|-----|---------|-------------|
| `ignore` | `[]` | Skip these packages in analysis |
| `exclude_dirs` | `node_modules`, `vendor`, `dist`, `build` | Directory basenames to skip |
| `exclude_files` | `[]` | Glob patterns to skip |
| `read_node_modules` | `false` | Scan `node_modules` for imports |
| `lock_file` | `true` | Enable lock file analysis |

## Architecture

```
depx
├── cmd/
│   ├── root.go                  # Root command
│   ├── scan.go                  # Scan subcommand
│   ├── surface.go               # Surface area analysis
│   └── config.go                # Config loading helper
├── internal/
│   ├── analyzer/                # Scan orchestration
│   ├── config/                  # .depx.yml parsing
│   ├── filter/                  # File exclusion rules
│   ├── lockfile/                # Lock file parsers
│   ├── manifest/                # Manifest parsers (npm/go/cargo/pip)
│   ├── report/                  # Terminal output
│   ├── surface/                 # Surface area analysis
│   └── usage/                   # Per-language import analyzers
├── tests/                       # Integration tests
└── testdata/                    # Fixture projects
```

## Technical Details

- **Language**: Go 1.26+
- **CLI**: cobra
- **Output**: fatih/color
- **Config**: gopkg.in/yaml.v3
- **Detection**: Regex + state-machine comment/string filtering

**Core flow:**

1. Detect project type (priority: npm → go → cargo → pip)
2. Parse manifest for declared dependencies
3. Walk source files and extract imports
4. Match declarations against usage
5. Optionally parse lock file for transitive dependencies
6. Generate terminal report

**Manifest detection:** If multiple manifest files exist (e.g. `package.json` + `go.mod`), npm takes priority and only npm dependencies are analyzed.

## Limitations

- **Usage detection** is static analysis — dynamic imports, reflection, and code generation may cause false positives/negatives
- **Scan** checks direct declarations in manifest (`dependencies` + `devDependencies`); **surface** defaults to runtime `dependencies` only
- **Indirect dependency lists** (`scan --indirect`) come from lock files; usage is not traced through transitive deps
- **Shared indirect graph** (`surface --indirect`) is available for npm `package-lock.json` only
- **Workspaces** — only the root manifest is analyzed, not per-package workspace members
- **npm `@types/*`** — tracked separately in scan; excluded from surface; typically not directly imported in source
- **Go `// indirect`** — excluded from manifest dependency list
- **Python** — package names may differ from import names (e.g. `Pillow` → `import PIL`)
- **exclude_dirs** matches directory **basename** only, not full paths

## License

GPLv3 — See [LICENSE](LICENSE) for details.
