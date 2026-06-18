# depx

Dependency Efficiency Analyzer - 依赖效率分析工具

Detect unused dependencies in your project and help identify dependency waste.

检测项目中声明但未使用的依赖，帮助开发者识别依赖浪费。

---

## Features / 功能

- Scan project dependency manifest files / 扫描项目依赖声明文件
- Analyze actual usage in source code / 分析源码中的实际使用情况
- Report unused dependencies / 输出未使用的依赖列表

## Supported / 支持范围

| Package Manager | Manifest | Source Files |
|----------------|----------|--------------|
| npm | package.json | .js, .ts, .jsx, .tsx, .mjs, .cjs, .vue, .svelte |
| Go | go.mod | .go |

## Installation / 安装

### Download from Release / 从 Release 下载

1. Go to [Releases](https://github.com/depx/depx/releases) and download the binary for your platform
2. Place the binary in any directory in your system PATH to use the `depx` command

前往 [Releases](https://github.com/depx/depx/releases) 下载对应平台的可执行文件，将其放到系统 PATH 中的任一目录即可使用。

Check current PATH / 查看当前 PATH：

```bash
# Windows
echo %PATH%

# Linux / macOS
echo $PATH
```

Or run with full path / 也可使用完整路径运行：

```bash
# Windows
C:\Users\xxx\Downloads\depx.exe scan

# Linux / macOS
~/Downloads/depx scan
```

### Build from Source / 从源码构建

```bash
git clone https://github.com/depx/depx.git
cd depx
go build -ldflags="-s -w" -o depx .
```

### Via go install / 通过 go install 安装

```bash
go install github.com/depx/depx@latest
```

## Usage / 使用方法

```bash
# Scan current directory / 扫描当前目录
depx scan

# Scan a specific directory / 扫描指定目录
depx scan /path/to/project
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

## Architecture / 架构

```
depx
├── cmd/                    # CLI entry / CLI 入口
│   ├── root.go            # Root command / 根命令
│   └── scan.go            # Scan subcommand / scan 子命令
├── internal/
│   ├── analyzer/          # Analyzer - orchestrates scanning / 分析器 - 协调扫描流程
│   │   └── unused_test.go # Integration tests / 集成测试
│   ├── manifest/          # Manifest parsing - reads dependency declarations / 清单解析
│   │   ├── npm.go         # package.json parser / package.json 解析
│   │   ├── gomod.go       # go.mod parser / go.mod 解析
│   │   └── manifest_test.go # Manifest parsing tests / 清单解析测试
│   ├── usage/             # Usage analysis - detects code references / 使用分析
│   │   ├── js.go          # JS/TS import analysis / JS/TS import 分析
│   │   ├── golang.go      # Go import analysis / Go import 分析
│   │   ├── js_test.go     # JS analyzer unit tests / JS 分析器单元测试
│   │   ├── golang_test.go # Go analyzer unit tests / Go 分析器单元测试
│   │   └── boundary_test.go # Boundary condition tests / 边界条件测试
│   └── report/            # Report generation - formatted output / 报告生成
├── testdata/              # Test fixtures / 测试夹具数据
│   ├── npm-project/       # npm basic project / npm 基础项目
│   ├── npm-complex/       # npm complex project (Vue/TS/hooks) / npm 复杂项目
│   ├── go-project/        # Go basic project / Go 基础项目
│   ├── go-complex/        # Go complex project (multi-package) / Go 复杂项目
│   ├── edge-all-used/     # Edge: all used / 边界：全部使用
│   ├── edge-none-used/    # Edge: none used / 边界：全部未使用
│   ├── edge-no-source/    # Edge: no source files / 边界：无源码
│   └── real-npm/          # npm real-world simulation / npm 真实场景模拟
└── main.go
```

## Technical Details / 技术实现

- **Language / 语言**: Go
- **CLI Framework / CLI 框架**: cobra
- **Colored Output / 输出着色**: fatih/color
- **Dependency Detection / 依赖检测**: Regex matching + state machine comment filtering / 正则匹配 + 状态机注释过滤

Core flow / 核心流程：

1. Detect project type (npm/go) / 检测项目类型
2. Parse manifest file to get dependency list / 解析清单文件获取依赖列表
3. Walk source files to extract import statements / 遍历源码文件提取 import 语句
4. Filter comments and string literals / 过滤注释和字符串字面量
5. Match dependency declarations with actual usage / 匹配依赖声明与实际使用
6. Generate report / 生成报告

## Limitations / 限制

- Only detects direct dependencies, does not analyze transitive dependencies / 仅检测直接依赖，不分析传递依赖
- In npm projects, `@types/*` packages always show as unused (auto-loaded by TypeScript compiler) / npm 项目中 `@types/*` 包始终显示为未使用
- In Go projects, dependencies marked `// indirect` are automatically excluded / Go 项目中自动排除 `// indirect` 标记的间接依赖

## License / 许可证

GPLv3 - See [LICENSE](LICENSE) for details.
