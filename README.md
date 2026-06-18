# depx

Dependency Efficiency Analyzer - 依赖效率分析工具

Detect unused dependencies in your project, analyze dependency efficiency and surface area.

检测项目中声明但未使用的依赖，分析依赖效率与影响面，帮助开发者识别依赖浪费。

---

## Features / 功能

- **Unused Detection** — Scan project dependency manifest files and detect unused dependencies / 扫描项目依赖声明文件，检测未使用的依赖
- **Efficiency Analysis** — Analyze which exports of a dependency are actually used, calculate efficiency percentage / 分析依赖的哪些导出被实际使用，计算效率百分比
- **Surface Area Analysis** — Analyze how widely a dependency is used across the project, assess criticality / 分析依赖在项目中的使用广度，评估关键度
- **Lock File Analysis** — Parse lock files to get accurate dependency versions and detect indirect dependencies / 解析 Lock File 获取准确的依赖版本并检测间接依赖
- **Monorepo Support** — Detect and scan npm workspaces, merge results across sub-projects / 检测并扫描 npm workspaces，合并子项目结果
- **Configuration** — Customize ignore rules, exclude directories via `.depx.yml` / 通过 `.depx.yml` 自定义忽略规则、排除目录

## Supported / 支持范围

| Package Manager | Manifest | Lock File | Source Files |
|----------------|----------|-----------|--------------|
| npm | package.json | package-lock.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | go.sum | .go |
| Rust | Cargo.toml | Cargo.lock | .rs |
| Python | requirements.txt | — | .py |

## Installation / 安装

### 首次设置（仅需一次）

Windows 默认禁止运行 PowerShell 脚本。首次使用前，请以**管理员身份**运行 PowerShell 并执行：

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

输入 `Y` 确认。此设置永久生效，之后可直接运行 `.ps1` 脚本。

### 安装 depx

```powershell
# 克隆仓库
git clone https://github.com/mukunjin/depx.git
cd depx

# 构建
go build -ldflags="-s -w -X github.com/mukunjin/depx/cmd.Version=v0.2.0" .

# 安装
.\install.ps1
```

### 卸载

```powershell
.\install.ps1 -Uninstall
```

脚本会自动完成：
- 将 `depx.exe` 复制到 `%LOCALAPPDATA%\depx`
- 添加到用户 PATH
- 提示重启终端使配置生效

## Usage / 使用方法

### Scan / 扫描

```bash
# Scan current directory / 扫描当前目录
depx scan

# Scan a specific directory / 扫描指定目录
depx scan C:\path\to\project

# Scan with custom config / 使用自定义配置扫描
depx scan --config C:\path\to\.depx.yml

# Show help / 显示帮助
depx --help

# Show version / 显示版本
depx --version
```

Example output / 输出示例：

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

### Efficiency Analysis / 效率分析

```bash
# Analyze dependency efficiency / 分析依赖效率
depx efficiency

# Analyze specific dependency / 分析特定依赖
depx efficiency lodash
```

Example output / 输出示例：

```
  Dependency Efficiency
--------------------------
  Package: lodash
  Functions Used: 1 (debounce)
  Estimated Exports: 30
  Efficiency: 3.3%
  Recommendation: Consider native implementation
```

### Surface Area Analysis / 影响面分析

```bash
# Analyze dependency surface area / 分析依赖影响面
depx surface
```

Example output / 输出示例：

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

### Monorepo Support / Monorepo 支持

```bash
# Scan monorepo workspaces / 扫描 monorepo 工作区
depx monorepo
```

Example output / 输出示例：

```
  Monorepo Summary
--------------------------
  Workspaces: 5
  Total Dependencies: 234
  Used: 189
  Unused: 45
```

## Architecture / 架构

```
depx
├── cmd/                          # CLI entry / CLI 入口
│   ├── efficiency.go            # Efficiency analysis command / 效率分析命令
│   ├── monorepo.go              # Monorepo support command / Monorepo 支持命令
│   ├── root.go                  # Root command / 根命令
│   ├── scan.go                  # Scan subcommand / scan 子命令
│   └── surface.go               # Surface area analysis command / 影响面分析命令
├── internal/
│   ├── analyzer/                # Analyzer - orchestrates scanning / 分析器 - 协调扫描流程
│   │   ├── monorepo.go          # Monorepo detection / Monorepo 检测
│   │   ├── monorepo_test.go     # Monorepo tests / Monorepo 测试
│   │   ├── unused.go            # Core scanning logic / 核心扫描逻辑
│   │   └── unused_test.go       # Analyzer tests / 分析器测试
│   ├── config/                  # Configuration / 配置管理
│   │   ├── config.go            # .depx.yml parsing / 配置文件解析
│   │   └── config_test.go       # Config tests / 配置测试
│   ├── efficiency/              # Efficiency analysis / 效率分析
│   │   ├── efficiency.go        # Core logic / 核心逻辑
│   │   ├── efficiency_test.go   # Efficiency tests / 效率测试
│   │   ├── go_export.go         # Go export extraction / Go 导出提取
│   │   └── js_export.go         # JS/TS export extraction / JS/TS 导出提取
│   ├── lockfile/                # Lock file parsing / Lock File 解析
│   │   ├── lockfile.go          # Unified interface / 统一接口
│   │   └── lockfile_test.go     # Lockfile tests / Lockfile 测试
│   ├── manifest/                # Manifest parsing / 清单解析
│   │   ├── cargo.go             # Cargo.toml parser / Cargo.toml 解析
│   │   ├── cargo_test.go        # Cargo tests / Cargo 测试
│   │   ├── gomod.go             # go.mod parser / go.mod 解析
│   │   ├── manifest.go          # Manifest interface / 清单接口
│   │   ├── manifest_test.go     # Manifest tests / 清单测试
│   │   ├── npm.go               # package.json parser / package.json 解析
│   │   ├── pip.go               # requirements.txt parser / requirements.txt 解析
│   │   └── pip_test.go          # Pip tests / Pip 测试
│   ├── report/                  # Report generation / 报告生成
│   │   └── terminal.go          # Terminal output / 终端输出
│   ├── surface/                 # Surface area analysis / 影响面分析
│   │   ├── surface.go           # Core logic / 核心逻辑
│   │   └── surface_test.go      # Surface tests / 影响面测试
│   └── usage/                   # Usage analysis / 使用分析
│       ├── boundary_test.go     # Boundary condition tests / 边界条件测试
│       ├── golang.go            # Go import analysis / Go import 分析
│       ├── golang_test.go       # Go analyzer tests / Go 分析器测试
│       ├── js.go                # JS/TS import analysis / JS/TS import 分析
│       ├── js_test.go           # JS analyzer tests / JS 分析器测试
│       ├── python.go            # Python import analysis / Python import 分析
│       ├── python_test.go       # Python analyzer tests / Python 分析器测试
│       ├── rust.go              # Rust use analysis / Rust use 分析
│       ├── rust_test.go         # Rust analyzer tests / Rust 分析器测试
│       └── usage.go             # Analyzer interface / 分析器接口
├── tests/                       # Integration tests / 集成测试
│   └── integration_test.go      # End-to-end tests / 端到端测试
├── testdata/                    # Test fixtures / 测试夹具数据
│   ├── edge-all-used/           # Edge: all used / 边界：全部使用
│   │   ├── index.js
│   │   └── package.json
│   ├── edge-no-source/          # Edge: no source files / 边界：无源码
│   │   └── package.json
│   ├── edge-none-used/          # Edge: none used / 边界：全部未使用
│   │   ├── index.js
│   │   └── package.json
│   ├── go-complex/              # Go complex project / Go 复杂项目
│   │   ├── handlers/
│   │   │   ├── handlers.go
│   │   │   └── handlers_test.go
│   │   ├── go.mod
│   │   └── main.go
│   ├── go-project/              # Go basic project / Go 基础项目
│   │   ├── go.mod
│   │   └── main.go
│   ├── npm-complex/             # npm complex project / npm 复杂项目
│   │   ├── src/
│   │   │   ├── __tests__/
│   │   │   │   └── index.test.ts
│   │   │   ├── hooks/
│   │   │   │   └── useApi.ts
│   │   │   ├── Component.vue
│   │   │   └── index.ts
│   │   └── package.json
│   ├── npm-project/             # npm basic project / npm 基础项目
│   │   ├── index.js
│   │   └── package.json
│   ├── python-project/          # Python project / Python 项目
│   │   ├── main.py
│   │   └── requirements.txt
│   ├── real-npm/                # npm real-world simulation / npm 真实场景模拟
│   │   ├── src/
│   │   │   ├── utils/
│   │   │   │   ├── api.js
│   │   │   │   └── helpers.js
│   │   │   ├── index.js
│   │   │   └── server.js
│   │   └── package.json
│   └── rust-project/            # Rust project / Rust 项目
│       ├── Cargo.toml
│       └── main.rs
├── .gitignore                   # Git ignore rules / Git 忽略规则
├── install.ps1                  # Windows install script / Windows 安装脚本
├── LICENSE                      # License file / 许可证文件
├── main.go                      # Entry point / 入口文件
├── README.md                    # Documentation / 文档
├── go.mod                       # Go module definition / Go 模块定义
└── go.sum                       # Go dependencies checksum / Go 依赖校验和
```

## Configuration / 配置

Create `.depx.yml` in your project root / 在项目根目录创建 `.depx.yml`：

```yaml
# Ignore specific dependencies / 忽略特定依赖
ignore:
  - "@types/node"
  - "typescript"

# Exclude directories / 排除目录
exclude_dirs:
  - "vendor"
  - "dist"
  - "node_modules"

# Exclude file patterns / 排除文件模式
exclude_files:
  - "*.test.js"
  - "*.spec.ts"

# Read node_modules for precise analysis / 读取 node_modules 进行精确分析
read_node_modules: false

# Enable lock file analysis / 启用 Lock File 分析
lock_file: true
```

## Technical Details / 技术实现

- **Language / 语言**: Go
- **CLI Framework / CLI 框架**: cobra
- **Colored Output / 输出着色**: fatih/color
- **YAML Parsing / YAML 解析**: gopkg.in/yaml.v3
- **Dependency Detection / 依赖检测**: Regex matching + state machine comment filtering / 正则匹配 + 状态机注释过滤

Core flow / 核心流程：

1. Detect project type (npm/go/cargo/pip) / 检测项目类型
2. Parse manifest file to get dependency list / 解析清单文件获取依赖列表
3. Parse lock file if available / 解析 Lock File（如果可用）
4. Load configuration from `.depx.yml` / 从 `.depx.yml` 加载配置
5. Walk source files to extract import statements / 遍历源码文件提取 import 语句
6. Filter comments and string literals / 过滤注释和字符串字面量
7. Match dependency declarations with actual usage / 匹配依赖声明与实际使用
8. Analyze efficiency and surface area / 分析效率和影响面
9. Generate report / 生成报告

## Limitations / 限制

- Only detects direct dependencies, does not analyze transitive dependencies / 仅检测直接依赖，不分析传递依赖
- In npm projects, `@types/*` packages always show as unused (auto-loaded by TypeScript compiler) / npm 项目中 `@types/*` 包始终显示为未使用
- In Go projects, dependencies marked `// indirect` are automatically excluded / Go 项目中自动排除 `// indirect` 标记的间接依赖
- Python package names may not match import names (e.g., `pip install Pillow` → `import PIL`) / Python 包名可能与 import 名不一致
- Monorepo support is limited to npm workspaces / Monorepo 支持仅限于 npm workspaces

## License / 许可证

GPLv3 - See [LICENSE](LICENSE) for details.
