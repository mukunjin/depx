# depx

[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-GPLv3-blue?style=flat-square)](./LICENSE)

Dependency Analyzer - Detect unused dependencies in your project and analyze dependency surface area.

**中文文档**: [README_zh.md](README_zh.md)

---

## Features

- **Unused Detection** — Scan project dependency manifest files and detect unused dependencies
- **Surface Area Analysis** — Analyze how widely a dependency is used across the project, assess criticality
- **Lock File Analysis** — Parse lock files to get accurate dependency versions and detect indirect dependencies
- **Configuration** — Customize ignore rules, exclude directories via `.depx.yml`

## Supported

| Package Manager | Manifest | Lock File | Source Files |
|----------------|----------|-----------|--------------|
| npm | package.json | package-lock.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | go.sum | .go |
| Rust | Cargo.toml | Cargo.lock | .rs |
| Python | requirements.txt | — | .py |

## Installation

### First-time Setup (Windows only)

Windows blocks PowerShell scripts by default. Run PowerShell as **Administrator** and execute:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Type `Y` to confirm. This setting persists.

### Install depx

```powershell
# Clone repository
git clone https://github.com/mukunjin/depx.git
cd depx

# Build (automatically gets version from Git tag)
.\build.ps1

# Install
.\install.ps1
```

### Uninstall

```powershell
.\install.ps1 -Uninstall
```

The script will:
- Copy `depx.exe` to `%LOCALAPPDATA%\depx`
- Add to user PATH
- Prompt to restart terminal

### Version Management

The version number is controlled by the following places:

| Location | Role | When it takes effect |
|----------|------|---------------------|
| Git tag | Primary source | When building via `.\build.ps1` |
| `cmd/root.go` line 8 | Fallback (`dev`) | When running `go build` directly without `-ldflags` |
| `build.ps1` | Reads git tag, injects via `-ldflags` | Every time you run `.\build.ps1` |
| `install.ps1` line 139 | Reads version from binary (`depx --version`) | During installation verification |

**How it works:**
1. `build.ps1` runs `git describe --tags --abbrev=0` to get the latest Git tag
2. The tag is injected into the binary via `-ldflags="-X github.com/mukunjin/depx/cmd.Version=<tag>"`
3. `cmd/root.go` provides a fallback value (`dev`) when no `-ldflags` is used
4. `install.ps1` verifies the installed binary by reading its version

**How to release a new version:**
1. Create a Git tag: `git tag v0.3.0`
2. Push the tag: `git push origin v0.3.0`
3. Run `.\build.ps1` to build with the new version
4. Run `.\install.ps1` to install

**Verify current version:**
```powershell
depx --version
```

## Usage

### Scan

```bash
# Scan current directory
depx scan

# Scan a specific directory
depx scan C:\path\to\project

# Scan with custom config
depx scan --config C:\path\to\.depx.yml

# Show indirect dependencies details (default: only show count)
depx scan --indirect
depx scan -i

# Show help
depx --help

# Show version
depx --version
```

Example output:

```
  Project Summary
--------------------------
  Path:            /path/to/project
  Package Manager: npm
  Dependencies:    12
  Used:            7
  Unused:          5

  Unused Dependencies
--------------------------
  [x] moment
  [x] chalk
  [x] typescript
```

### Surface Area Analysis

```bash
# Analyze dependency surface area
depx surface
```

Example output:

```
  Dependency Surface Area
--------------------------
  axios
    Files: 43
    Modules: 8
    References: 182
    Criticality: High

  chalk
    Files: 4
    Modules: 1
    References: 6
    Criticality: Low
```

## Architecture

```
depx
├── cmd/
│   ├── root.go                  # Root command
│   ├── root_test.go
│   ├── scan.go                  # Scan subcommand
│   ├── scan_test.go
│   ├── surface.go               # Surface area analysis command
│   └── surface_test.go
├── internal/
│   ├── analyzer/
│   │   ├── unused.go            # Core scanning logic
│   │   └── unused_test.go
│   ├── config/
│   │   ├── config.go            # .depx.yml parsing
│   │   └── config_test.go
│   ├── filter/
│   │   ├── file.go              # File/directory exclusion rules
│   │   └── file_test.go
│   ├── lockfile/
│   │   ├── lockfile.go          # Unified interface
│   │   └── lockfile_test.go
│   ├── manifest/
│   │   ├── cargo.go             # Cargo.toml parser
│   │   ├── cargo_test.go
│   │   ├── gomod.go             # go.mod parser
│   │   ├── manifest.go          # Manifest interface
│   │   ├── manifest_test.go
│   │   ├── npm.go               # package.json parser
│   │   ├── pip.go               # requirements.txt parser
│   │   └── pip_test.go
│   ├── report/
│   │   ├── terminal.go          # Terminal output
│   │   └── terminal_test.go
│   ├── surface/
│   │   ├── surface.go           # Core logic
│   │   └── surface_test.go
│   └── usage/
│       ├── boundary_test.go     # Boundary condition tests
│       ├── golang.go            # Go import analysis
│       ├── golang_test.go
│       ├── js.go                # JS/TS import analysis
│       ├── js_test.go
│       ├── python.go            # Python import analysis
│       ├── python_test.go
│       ├── rust.go              # Rust use analysis
│       ├── rust_test.go
│       └── usage.go             # Analyzer interface
├── tests/
│   ├── integration_test.go      # End-to-end tests
│   └── helpers/
│       └── helpers.go           # Test helper functions
├── testdata/
│   ├── edge-all-used/
│   │   ├── index.js
│   │   └── package.json
│   ├── edge-empty/
│   │   ├── index.js
│   │   └── package.json
│   ├── edge-large/
│   │   ├── index.js
│   │   └── package.json
│   ├── edge-no-source/
│   │   └── package.json
│   ├── edge-none-used/
│   │   ├── index.js
│   │   └── package.json
│   ├── edge-special-chars/
│   │   ├── index.ts
│   │   └── package.json
│   ├── go-complex/
│   │   ├── handlers/
│   │   │   ├── handlers.go
│   │   │   └── handlers_test.go
│   │   ├── go.mod
│   │   └── main.go
│   ├── go-project/
│   │   ├── go.mod
│   │   └── main.go
│   ├── npm-complex/
│   │   ├── src/
│   │   │   ├── __tests__/
│   │   │   │   └── index.test.ts
│   │   │   ├── hooks/
│   │   │   │   └── useApi.ts
│   │   │   ├── Component.vue
│   │   │   └── index.ts
│   │   └── package.json
│   ├── npm-project/
│   │   ├── index.js
│   │   └── package.json
│   ├── python-complex/
│   │   ├── app.py
│   │   ├── database.py
│   │   ├── models.py
│   │   └── requirements.txt
│   ├── python-project/
│   │   ├── main.py
│   │   └── requirements.txt
│   ├── real-npm/
│   │   ├── src/
│   │   │   ├── utils/
│   │   │   │   ├── api.js
│   │   │   │   └── helpers.js
│   │   │   ├── index.js
│   │   │   └── server.js
│   │   └── package.json
│   ├── rust-complex/
│   │   ├── src/
│   │   │   ├── handlers.rs
│   │   │   └── main.rs
│   │   └── Cargo.toml
│   ├── rust-project/
│   │   ├── Cargo.toml
│   │   └── main.rs
│   ├── config-project/
│   │   ├── .depx.yml
│   │   ├── index.js
│   │   └── package.json
│   ├── complex-mixed/
│   │   ├── Cargo.toml
│   │   ├── go.mod
│   │   ├── package.json
│   │   ├── requirements.txt
│   │   ├── index.js
│   │   ├── main.go
│   │   ├── main.py
│   │   └── lib.rs
│   ├── complex-npm-workspaces/
│   │   ├── package.json
│   │   └── packages/
│   │       ├── core/
│   │       │   ├── package.json
│   │       │   └── index.ts
│   │       └── utils/
│   │           ├── package.json
│   │           └── index.ts
│   ├── complex-cargo-workspaces/
│   │   ├── Cargo.toml
│   │   ├── src/
│   │   │   └── main.rs
│   │   └── crates/
│   │       ├── core/
│   │       │   ├── Cargo.toml
│   │       │   └── src/
│   │       │       └── lib.rs
│   │       └── utils/
│   │           ├── Cargo.toml
│   │           └── src/
│   │               └── lib.rs
│   ├── error-corrupted-lockfile/
│   │   ├── index.js
│   │   ├── package.json
│   │   └── package-lock.json
│   ├── error-invalid-json/
│   │   └── package.json
│   └── error-invalid-toml/
│       ├── Cargo.toml
│       └── main.rs
├── .gitignore
├── build.ps1                    # Build script (auto version from Git tag)
├── install.ps1                  # Windows install script
├── LICENSE
├── main.go                      # Entry point
├── README.md                    # Documentation (English)
├── README_zh.md                 # Documentation (Chinese)
├── go.mod                       # Go module definition
└── go.sum                       # Go dependencies checksum
```

## Configuration

Create `.depx.yml` in your project root:

```yaml
# Ignore specific dependencies
ignore:
  - "@types/node"
  - "typescript"

# Exclude directories
exclude_dirs:
  - "vendor"
  - "dist"
  - "node_modules"

# Exclude file patterns
exclude_files:
  - "*.test.js"
  - "*.spec.ts"

# Read node_modules for precise analysis
read_node_modules: false

# Enable lock file analysis
lock_file: true
```

## Technical Details

- **Language**: Go
- **CLI Framework**: cobra
- **Colored Output**: fatih/color
- **YAML Parsing**: gopkg.in/yaml.v3
- **Dependency Detection**: Regex matching + state machine comment filtering

Core flow:

1. Detect project type (npm/go/cargo/pip)
2. Parse manifest file to get dependency list
3. Parse lock file if available
4. Load configuration from `.depx.yml`
5. Walk source files to extract import statements
6. Filter comments and string literals
7. Match dependency declarations with actual usage
8. Analyze surface area
9. Generate report

## Limitations

- Only detects direct dependencies, does not analyze transitive dependencies
- In npm projects, `@types/*` packages always show as unused (auto-loaded by TypeScript compiler)
- In Go projects, dependencies marked `// indirect` are automatically excluded
- Python package names may not match import names (e.g., `pip install Pillow` → `import PIL`)

## License

GPLv3 - See [LICENSE](LICENSE) for details.
